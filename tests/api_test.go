package tests

import (
	"auxstream/api"
	"auxstream/db"
	fs "auxstream/file_system"
	"context"
	"fmt"
	"github.com/imroc/req"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v2"
	"github.com/stretchr/testify/require"
	"log"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

var cwd, _ = os.Getwd()
var testDataPath = filepath.Join(cwd, "testdata")
var mockConn pgxmock.PgxConnIface
var router = api.SetupRouter()

func setupTest(t *testing.T) func(t *testing.T) {
	var err error
	mockConn, err = pgxmock.NewConn()
	if err != nil {
		log.Fatalf("Failed to set up mock database connection: %v", err)
	}
	db.DAO = db.NewWithMockConn(mockConn)

	return tearDownTest
}

func tearDownTest(t *testing.T) {
	_ = mockConn.Close(context.Background())
}

func TestHTTPAddTrack(t *testing.T) {
	teardown := setupTest(t)
	defer teardown(t)

	columns := []string{"id", "created_at"}
	mockConn.ExpectBegin()
	mockConn.ExpectQuery("INSERT INTO auxstream.artists").
		WithArgs("Sample Artist").
		WillReturnRows(pgxmock.NewRows(columns).AddRow(1, time.Now()))
	mockConn.ExpectQuery("INSERT INTO auxstream.tracks").
		WithArgs("Sample Title", 1, "auffdadf").
		WillReturnRows(pgxmock.NewRows(columns).AddRow(1, time.Now()))
	mockConn.ExpectCommit()
	// mock real store with test store
	fs.Store = fs.NewStore(os.TempDir())
	tserver := httptest.NewServer(router)

	defer tserver.Close()

	title := "Sample Title"
	artist := "Sample Artist"
	audioFilePath := filepath.Join(testDataPath, "audio", "audio.mp3")
	file, err := os.Open(audioFilePath)

	require.NoError(t, err)

	// Create the request body
	body := req.Param{
		"title":  title,
		"artist": artist,
	}

	post, err := req.Post(tserver.URL+"/upload_track", body, req.FileUpload{
		FieldName: "audio",
		File:      file,
		FileName:  "audio",
	})
	require.Equal(t, post.Response().StatusCode, 200)
	require.Equal(t, fs.Store.Writes(), 1)
	data := &map[string]interface{}{}
	err = post.ToJSON(data)
	require.NoError(t, err)
	fmt.Println("response body: ", data)
}

func TestHTTPSearchByArtist(t *testing.T) {
	teardown := setupTest(t)
	defer teardown(t)

	columns := []string{"id", "title", "artist_id", "file", "created_at"}
	mockConn.ExpectQuery(`
	SELECT t.id, t.title, t.artist_id, t.file, t.created_at
	`).
		WithArgs("Hike").
		WillReturnRows(pgxmock.NewRows(columns).
			AddRow(1, "Title", 1, "Test file", time.Now()).
			AddRow(1, "Title", 1, "Test file", time.Now()).
			AddRow(1, "Title", 1, "Test file", time.Now()))
	tserver := httptest.NewServer(router)
	defer tserver.Close()

	resp, err := req.Get(tserver.URL + "/search?artist=Hike")

	require.NoError(t, err)
	require.Equal(t, resp.Response().StatusCode, 200)
	data := &map[string]interface{}{}
	err = resp.ToJSON(data)
	require.NoError(t, err)
	fmt.Println("response body: ", data)
}

func TestHTTPTrackUploadBatch(t *testing.T) {

	teardown := setupTest(t)
	defer teardown(t)
	testRecordCnt := 30
	mockConn.ExpectCopyFrom(pgx.Identifier{"auxstream", "tracks"}, []string{"title", "artist_id", "file"}).
		WillReturnResult(int64(testRecordCnt))
	fs.Store = fs.NewStore(os.TempDir())
	var err error
	var trackFiles []req.FileUpload

	artistId := 2
	audioFilePath := filepath.Join(testDataPath, "audio", "audio.mp3")

	tserver := httptest.NewServer(router)

	for i := 0; i < testRecordCnt; i++ {
		file, err := os.Open(audioFilePath)
		require.NoError(t, err)
		trackFiles = append(trackFiles, req.FileUpload{
			FieldName: "track_files",
			File:      file,
			FileName:  "audio",
		})
	}

	formData := url.Values{}
	formData.Add("artist_id", strconv.Itoa(artistId))
	for i := 0; i < testRecordCnt; i++ {
		formData.Add("track_title", fmt.Sprintf("#%d", i))
	}

	post, err := req.Post(tserver.URL+"/upload_batch_track", formData, trackFiles)
	require.NoError(t, err)
	data := &map[string]interface{}{}
	err = post.ToJSON(data)
	require.NoError(t, err)
	fmt.Println("response body: ", data)
	require.Equal(t, post.Response().StatusCode, 200)
}
