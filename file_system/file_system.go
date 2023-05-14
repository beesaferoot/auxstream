package filesystem

import (
	"io"
)

// A generic interface for file storage operations 
type FileSystem interface {
	Save(file_name string, raw []byte) error
	Read(file_name string) (file File, err error)
	Reads() int
	Writes() int
	BulkSave(list_of_file_names []string, list_of_raw [][]byte )
	Remove(file_name string) error 
}

// encasulate file representation from other packages 
type File interface {
	Name() string
	Size() int64
	io.ReadWriter
}
