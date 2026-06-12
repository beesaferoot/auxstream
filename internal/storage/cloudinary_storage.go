package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api"
	_ "github.com/cloudinary/cloudinary-go/v2/api/admin"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

// CloudinaryStore is a FileSystem backed by Cloudinary. Assets are uploaded as the
// "video" resource type (Cloudinary's category for audio too). Save returns the
// asset's secure HTTPS URL as the identifier, which Read fetches over HTTP. Remove,
// however, expects the Cloudinary public ID rather than that URL.
type CloudinaryStore struct {
	cloudinaryInstance *cloudinary.Cloudinary
	mu                 sync.Mutex
	uploads            int
	downloads          int
}

// NewCloudinaryStore builds a store from a cloudinary:// URL carrying the cloud name
// and API credentials (the same format as the CLOUDINARY_URL environment variable).
func NewCloudinaryStore(url string) (*CloudinaryStore, error) {
	cld, err := cloudinary.NewFromURL(url)
	if err != nil {
		return nil, err
	}
	return &CloudinaryStore{
		cloudinaryInstance: cld,
	}, nil
}

func (cld *CloudinaryStore) Reads() int {
	return cld.downloads
}

func (cld *CloudinaryStore) Writes() int {
	return cld.uploads
}

func (cld *CloudinaryStore) Save(raw []byte, ext string) (filename string, err error) {
	if len(raw) < 1 {
		return "", errors.New("empty file")
	}

	filename = genFileName(ext)

	freader := bytes.NewReader(raw)

	uploadRes, err := cld.cloudinaryInstance.Upload.Upload(
		context.Background(),
		freader,
		uploader.UploadParams{
			// Cloudinary public IDs carry no extension; it derives that from the asset.
			PublicID:       strings.TrimSuffix(filename, "."+ext),
			ResourceType:   "video",
			UniqueFilename: api.Bool(false),
			Overwrite:      api.Bool(true),
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to upload file, %v", err)
	}

	filename = uploadRes.SecureURL
	cld.mu.Lock()
	cld.uploads++
	cld.mu.Unlock()
	return filename, nil
}

func (cld *CloudinaryStore) Read(locationURL string) (File, error) {
	resp, err := http.Get(locationURL) //nolint:gosec // URL comes from Cloudinary upload result
	if err != nil {
		return nil, fmt.Errorf("failed to fetch asset: %w", err)
	}
	defer resp.Body.Close()

	// Stream the fetched asset into a local temp file and return that as the File;
	// the caller owns closing (and cleaning up) it.
	file, err := NewFile(os.TempDir() + genFileName("tmp"))
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}

	if _, err = io.Copy(file, resp.Body); err != nil {
		return nil, fmt.Errorf("failed to write asset to file: %w", err)
	}

	cld.mu.Lock()
	cld.downloads++
	cld.mu.Unlock()

	return file, nil
}

func (cld *CloudinaryStore) BulkSave(buf chan<- FileMeta, listOfFileMeta []FileMeta) {
	var wg sync.WaitGroup
	for _, fd := range listOfFileMeta {
		wg.Add(1)
		go func(raw []byte, title, ext string) {
			defer wg.Done()
			fileUrl, err := cld.Save(raw, ext)
			if err != nil {
				log.Printf("bulk save error: %s", err.Error())
				buf <- FileMeta{AudioTitle: title}
				return
			}
			buf <- FileMeta{Name: fileUrl, Content: raw, AudioTitle: title, Ext: ext}
		}(fd.Content, fd.AudioTitle, fd.Ext)
	}
	wg.Wait()
	close(buf)
}

func (cld *CloudinaryStore) Remove(locationUrl string) error {
	_, err := cld.cloudinaryInstance.Upload.Destroy(context.Background(), uploader.DestroyParams{PublicID: locationUrl})
	if err != nil {
		log.Printf("remove resource error: %s", err.Error())
		return err
	}
	return nil
}
