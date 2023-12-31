package tests

import (
	store "auxstream/file_system"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCreates3Store(t *testing.T) {
	var s3Store store.FileSystem = store.NewS3Store("test-bucket")
	require.Equal(t, 0, s3Store.Reads())
	require.Equal(t, 0, s3Store.Writes())
}
