package storage

import (
	"auxstream/config"
	"io"
)

// FileMeta carries one file through a BulkSave batch and its per-file result.
type FileMeta struct {
	Name       string // backend identifier from Save; empty if that file's save failed
	AudioTitle string // caller-supplied logical title, preserved verbatim across the round trip
	Content    []byte // raw bytes; the input on the way in, echoed back on success
	Ext        string // file extension without the leading dot, e.g. "mp3"
}

// FileSystem stores opaque blobs and addresses each by a backend-issued identifier.
// The string returned by Save is that identifier; Read and Remove consume the same
// value. Its form is backend-specific (a bare filename, an S3 URL, a Cloudinary URL),
// so callers must treat it as opaque and never construct or parse it themselves.
type FileSystem interface {
	// Save persists raw and returns the identifier under which it is now addressable.
	// ext sets the stored extension (defaulting to mp3 when empty). Empty input is
	// rejected rather than stored.
	Save(raw []byte, ext string) (filename string, err error)
	// Read fetches the blob named by an identifier from a prior Save. The returned
	// File holds an open handle the caller owns and must Close.
	Read(fileName string) (file File, err error)
	// Reads reports the number of successful Read calls over this store's lifetime.
	Reads() int
	// Writes reports the number of successful Save calls over this store's lifetime.
	Writes() int
	// BulkSave saves every entry concurrently and emits one FileMeta per entry on buf,
	// closing buf when done. A failed entry yields a FileMeta with only AudioTitle set,
	// so results arrive in completion order, not input order.
	BulkSave(buf chan<- FileMeta, listOfFileMeta []FileMeta)
	// Remove deletes the blob named by an identifier from a prior Save.
	Remove(fileName string) error
}

// File is an open handle to a stored blob, readable and writable in place.
// The caller that obtains one (e.g. from Read) owns it and is responsible for Close.
type File interface {
	// Name reports the handle's underlying path or identifier, which may differ
	// from the identifier passed to Read (e.g. a local temp path for remote backends).
	Name() string
	Size() int64
	io.ReadWriter
}

// SetFileStore swaps the package-level Store to the configured backend. Unrecognized
// or incompletely configured backends are ignored, leaving the local-disk default.
func SetFileStore(config config.Config) error {

	if config.FileStore == "s3" && config.S3bucket != "" {
		Store = NewS3Store(config.S3bucket)
	} else if config.FileStore == "cloudinary" && config.CloudinaryURL != "" {
		st, err := NewCloudinaryStore(config.CloudinaryURL)
		if err != nil {
			return err
		}
		Store = st
	}

	return nil
}
