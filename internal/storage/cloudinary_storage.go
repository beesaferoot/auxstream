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

// CloundinaryStore a FileSystem interface def over cloudinary
type CloudinaryStore struct {
	cloudinaryInstance *cloudinary.Cloudinary
	mu                 sync.Mutex
	uploads            int
	downloads          int
}

func NewCloudinaryStore(url string) (*CloudinaryStore, error) {
	// Start by creating a new instance of Cloudinary using CLOUDINARY_URL environment variable.
	// Alternatively you can use cloudinary.NewFromParams() or cloudinary.NewFromURL().
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

func (cld *CloudinaryStore) Save(raw []byte) (filename string, err error) {
	if len(raw) < 1 {
		return "", errors.New("empty file")
	}

	filename = genFileName()

	freader := bytes.NewReader(raw)

	uploadRes, err := cld.cloudinaryInstance.Upload.Upload(
		context.Background(),
		freader,
		uploader.UploadParams{
			PublicID:       strings.TrimSuffix(filename, ".mp3"),
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

func (cld *CloudinaryStore) Read(locationUrl string) (file File, err error) {
	resp, err := http.Get(locationUrl)

	file, err = NewFile(os.TempDir() + genFileName())

	if err != nil {
		return nil, fmt.Errorf("failed to get asset, %v", err)
	}

	_, err = io.Copy(file, resp.Body)

	if err == nil {
		cld.mu.Lock()
		cld.downloads++
		cld.mu.Unlock()
	}
	return file, nil
}

func (cld *CloudinaryStore) BulkSave(buf chan<- string, listOfRaw [][]byte) {
	var wg sync.WaitGroup
	for _, raw := range listOfRaw {
		wg.Add(1)
		go func(raw []byte) {
			defer wg.Done()
			fileUrl, err := cld.Save(raw)
			if err != nil {
				log.Printf("bulk save error: %s", err.Error())
				buf <- ""
				return
			}
			buf <- fileUrl
		}(raw)
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
