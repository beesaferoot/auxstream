package tests

import (
	"auxstream/api"
	"auxstream/db"
	fs "auxstream/file_system"
	"context"
	"fmt"
	"github.com/imroc/req"
	"github.com/pashagolub/pgxmock/v2"
	"github.com/stretchr/testify/require"
	"log"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

var cwd, _ = os.Getwd()
var testDataPath = filepath.Join(cwd, "testdata")
var mockConn pgxmock.PgxConnIface
var router = api.SetupRouter()

func TestMain(m *testing.M) {
	// Set up the test database
	var err error
	mockConn, err = pgxmock.NewConn()
	if err != nil {
		log.Fatalf("Failed to set up mock database connection: %v", err)
	}
	defer mockConn.Close(context.Background())

	db.DAO = db.NewWithMockConn(mockConn)
	// Run the tests
	code := m.Run()
	if err != nil {
		os.Exit(1)
	}
	// Exit the test
	os.Exit(code)
}

func TestHTTPAddTrack(t *testing.T) {
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
