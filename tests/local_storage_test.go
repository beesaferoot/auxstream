package tests

import (
	store "auxstream/file_system"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

var baseLocation, _ = os.Getwd()

func TestCreateStore(t *testing.T) {
	var lstore store.FileSystem = store.NewStore(baseLocation)
	require.Equal(t, 0, lstore.Reads())
	require.Equal(t, 0, lstore.Writes())
}

func TestSaveFile(t *testing.T) {
	var lstore store.FileSystem = store.NewStore(baseLocation)
	lstore.Save("hello_world_test.txt", []byte("hello world"))
	lstore.Save("second_entry.txt", []byte("second file content"))
	defer lstore.Remove("hello_world_test.txt")
	defer lstore.Remove("second_entry.txt")
	require.Equal(t, 2, lstore.Writes())
}

func TestReadFile(t *testing.T) {
	var lstore store.FileSystem = store.NewStore(baseLocation)
	write_bytes := []byte("hello world")
	lstore.Save("hello_world_test.txt", write_bytes)
	defer lstore.Remove("hello_world_test.txt")
	file, err := lstore.Read("hello_world_test.txt")
	require.NoError(t, err)
	read_bytes := make([]byte, len(write_bytes))
	_, err = file.Read(read_bytes)
	require.NoError(t, err)
	require.Equal(t, "hello world", string(read_bytes))
	require.Equal(t, 1, lstore.Reads())
}

func TestBulkSave(t *testing.T) {
	var lstore store.FileSystem = store.NewStore(baseLocation)
	groupfiles := [][]byte{}
	file_names := []string{}
	for i := 0; i < 10; i++ {
		groupfiles = append(groupfiles, []byte("hello world"))
		file_names = append(file_names, fmt.Sprintf("text_bulk_save_%d", i))
	}
	lstore.BulkSave(file_names, groupfiles)
	require.Equal(t, 10, lstore.Writes())
	for _, file_name := range file_names {
		err := lstore.Remove(file_name)
		require.NoError(t, err)
	}
}
