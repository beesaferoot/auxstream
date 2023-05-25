package filesystem

import (
	"io"
)

// FileSystem A generic interface for file storage operations
type FileSystem interface {
	Save(raw []byte) (filename string, err error)
	Read(fileName string) (file File, err error)
	Reads() int
	Writes() int
	BulkSave(buf chan<- string, listOfRaw [][]byte)
	Remove(fileName string) error
}

// File encapsulate file representation from other packages
type File interface {
	Name() string
	Size() int64
	io.ReadWriter
}
