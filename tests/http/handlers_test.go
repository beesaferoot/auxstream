package tests

import (
	"auxstream/internal/cache"
	"auxstream/internal/http"
	fs "auxstream/internal/storage"
	"database/sql"
	"fmt"
	"log"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/imroc/req"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var cwd, _ = os.Getwd()

// Fixtures live in tests/testdata, one level up from this package (tests/http).
var testDataPath = filepath.Join(cwd, "..", "testdata")
var mockDB *sql.DB
var sqlMock sqlmock.Sqlmock
var router *gin.Engine
var gormDB *gorm.DB

func setupTest(_ *testing.T) func(t *testing.T) {
	var err error

	// Create a mock sql.DB using go-sqlmock
	mockDB, sqlMock, err = sqlmock.New()
	if err != nil {
		log.Fatalf("Failed to create mock database: %v", err)
	}

	// Create a GORM DB instance using the mock sql.DB
	gormDB, err = gorm.Open(postgres.New(postgres.Config{
		Conn: mockDB,
	}), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to create GORM mock DB: %v", err)
	}

	mr, _ := miniredis.Run()
	opts := &redis.Options{
		Addr: mr.Addr(),
	}
	r := cache.NewRedis(opts)
	server := http.NewMockServer(gormDB, r)
	router = server.SetupRouter(true)

	return tearDownTest
}

func tearDownTest(_ *testing.T) {
	_ = mockDB.Close()
}

func TestHTTPAddTrack(t *testing.T) {
	teardown := setupTest(t)
	defer teardown(t)

	artistID := uuid.New()
	trackID := uuid.New()

	// Mock artist lookup
	sqlMock.ExpectQuery(`SELECT .* FROM "auxstream"\."artists"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "created_at", "updated_at"}).
			AddRow(artistID, "Hike", time.Now(), time.Now()))

	sqlMock.ExpectBegin()
	// Mock track creation
	sqlMock.ExpectQuery(`INSERT INTO "auxstream"\."tracks"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(trackID))
	sqlMock.ExpectCommit()

	// Mock real store with test store
	fs.Store = fs.NewLocalStore(os.TempDir())
	tserver := httptest.NewServer(router)
	defer tserver.Close()

	title := "Sample Title"
	audioFilePath := filepath.Join(testDataPath, "audio", "audio.mp3")
	file, err := os.Open(audioFilePath)

	require.NoError(t, err)

	// Create the request body
	body := req.Param{
		"title":     title,
		"artist_id": artistID.String(),
	}

	post, err := req.Post(tserver.URL+"/upload_track", body, req.FileUpload{
		FieldName: "audio",
		File:      file,
		FileName:  "audio.mp3",
	})
	require.Equal(t, 200, post.Response().StatusCode)
	require.Equal(t, fs.Store.Writes(), 1)
	data := &map[string]any{}
	err = post.ToJSON(data)
	require.NoError(t, err)
	// Ensure all expectations were met
	require.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestHTTPSearchByArtist(t *testing.T) {
	teardown := setupTest(t)
	defer teardown(t)

	artistID := uuid.New()

	// Expect Preloaded artist
	// Mock the JOIN query for artist search
	sqlMock.ExpectQuery(`SELECT .* FROM "auxstream"\."tracks" JOIN auxstream.artists`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "artist_id", "file", "created_at", "updated_at"}).
			AddRow(uuid.New(), "Title", artistID, "Test file", time.Now(), time.Now()).
			AddRow(uuid.New(), "Title", artistID, "Test file", time.Now(), time.Now()).
			AddRow(uuid.New(), "Title", artistID, "Test file", time.Now(), time.Now()))

	tserver := httptest.NewServer(router)
	defer tserver.Close()

	resp, err := req.Get(tserver.URL + "/search?artist=Hike")

	require.NoError(t, err)
	require.Equal(t, 200, resp.Response().StatusCode)
	data := &map[string]any{}
	err = resp.ToJSON(data)
	require.NoError(t, err)

	// Ensure all expectations were met
	require.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestHTTPTrackUploadBatch(t *testing.T) {
	teardown := setupTest(t)
	defer teardown(t)
	testRecordCnt := 30

	// Mock batch insert: CreateInBatches issues one INSERT ... RETURNING "id",
	// and GORM scans one returned id per row, so return a UUID for each track.
	insertedIDs := sqlmock.NewRows([]string{"id"})
	for range testRecordCnt {
		insertedIDs.AddRow(uuid.New())
	}
	sqlMock.ExpectBegin()
	sqlMock.ExpectQuery(`INSERT INTO "auxstream"\."tracks" \(.+\) VALUES .+ RETURNING "id"`).
		WillReturnRows(insertedIDs)
	sqlMock.ExpectCommit()

	fs.Store = fs.NewLocalStore(os.TempDir())
	var err error
	var trackFiles []req.FileUpload

	artistID := uuid.New()
	audioFilePath := filepath.Join(testDataPath, "audio", "audio.mp3")

	tserver := httptest.NewServer(router)

	for range testRecordCnt {
		file, err := os.Open(audioFilePath)
		require.NoError(t, err)
		trackFiles = append(trackFiles, req.FileUpload{
			FieldName: "track_files",
			File:      file,
			FileName:  "audio",
		})
	}

	formData := url.Values{}
	formData.Add("artist_id", artistID.String())
	for i := range testRecordCnt {
		formData.Add("track_titles", fmt.Sprintf("#%d", i))
	}

	post, err := req.Post(tserver.URL+"/upload_batch_track", formData, trackFiles)
	require.NoError(t, err)
	data := &map[string]any{}
	err = post.ToJSON(data)
	require.NoError(t, err)
	require.Equal(t, 200, post.Response().StatusCode)

	// Every uploaded file shares the filename "audio"; the bulk path must still
	// persist all of them by correlating on title, not filename.
	require.Equal(t, float64(testRecordCnt), (*data)["data"].(map[string]any)["rows"])

	// Ensure all expectations were met
	require.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestHTTPFetchTracks(t *testing.T) {
	teardown := setupTest(t)
	defer teardown(t)

	artistID := uuid.New()

	// Mock tracks query with pagination
	sqlMock.ExpectQuery(`SELECT (.+) FROM "auxstream"\."tracks"`).
		WithArgs(2).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "artist_id", "file", "created_at", "updated_at"}).
			AddRow(uuid.New(), "Title", artistID, "Test file", time.Now(), time.Now()).
			AddRow(uuid.New(), "Title", artistID, "Test file", time.Now(), time.Now()).
			AddRow(uuid.New(), "Title", artistID, "Test file", time.Now(), time.Now()))

	// GORM Preload("Artist") issues a follow-up query to hydrate the artists.
	sqlMock.ExpectQuery(`SELECT (.+) FROM "auxstream"\."artists"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "created_at", "updated_at"}).
			AddRow(artistID, "Hike", time.Now(), time.Now()))

	tserver := httptest.NewServer(router)
	defer tserver.Close()

	resp, err := req.Get(tserver.URL + "/tracks?pagesize=2&pagenumber=1")

	require.NoError(t, err)
	require.Equal(t, 200, resp.Response().StatusCode)
	data := &map[string]any{}
	err = resp.ToJSON(data)
	require.NoError(t, err)
	require.NotZero(t, data)
	resData := (*data)["data"]
	require.NotEmpty(t, resData)

	// Ensure all expectations were met
	require.NoError(t, sqlMock.ExpectationsWereMet())
}
