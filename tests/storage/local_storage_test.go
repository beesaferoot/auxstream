package tests

import (
	store "auxstream/internal/storage"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

var baseLocation, _ = os.Getwd()

func TestCreateStore(t *testing.T) {
	var lstore store.FileSystem = store.NewLocalStore(baseLocation)
	require.Equal(t, 0, lstore.Reads())
	require.Equal(t, 0, lstore.Writes())
}

func TestSaveFile(t *testing.T) {
	var lstore store.FileSystem = store.NewLocalStore(baseLocation)
	file1, _ := lstore.Save([]byte("hello world"))
	file2, _ := lstore.Save([]byte("second file content"))
	defer lstore.Remove(file1)
	defer lstore.Remove(file2)
	require.Equal(t, 2, lstore.Writes())
}

func TestReadFile(t *testing.T) {
	var lstore store.FileSystem = store.NewLocalStore(baseLocation)
	write_bytes := []byte("hello world")
	fileName, _ := lstore.Save(write_bytes)
	defer lstore.Remove(fileName)
	file, err := lstore.Read(fileName)
	require.NoError(t, err)
	read_bytes := make([]byte, len(write_bytes))
	_, err = file.Read(read_bytes)
	require.NoError(t, err)
	require.Equal(t, "hello world", string(read_bytes))
	require.Equal(t, 1, lstore.Reads())
}

func TestBulkSave(t *testing.T) {
	var lstore store.FileSystem = store.NewLocalStore(baseLocation)
	var groupfiles [][]byte
	var fileNames []string

	for range 10 {
		groupfiles = append(groupfiles, []byte("hello world"))
	}
	buf := make(chan string, len(groupfiles))
	lstore.BulkSave(buf, groupfiles)
	for fileName := range buf {
		fileNames = append(fileNames, fileName)
	}
	require.Equal(t, 10, lstore.Writes())
	require.Equal(t, 10, len(fileNames))
	for _, fileName := range fileNames {
		err := lstore.Remove(fileName)
		require.NoError(t, err)
	}
}
