package filesystem

import (
	"bytes"
	"context"
	"errors"
	"fmt"
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

func NewCloudinaryStore() (*CloudinaryStore, error) {
	// Start by creating a new instance of Cloudinary using CLOUDINARY_URL environment variable.
	// Alternatively you can use cloudinary.NewFromParams() or cloudinary.NewFromURL().
	cld, err := cloudinary.New()
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
	if len(raw) > 0 {
		return "", errors.New("empty file")
	}

	filename = genFileName()

	freader := bytes.NewReader(raw)

	uploadRes, err := cld.cloudinaryInstance.Upload.Upload(
		context.Background(),
		freader,
		uploader.UploadParams{
			PublicID:       filename,
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

func (cld *CloudinaryStore) Read(location string) (file File, err error) {
	//asset, err := cld.cloudinaryInstance.Admin.Asset(context.Background(), admin.AssetParams{
	//	AssetType: "audio",
	//	PublicID:  location,
	//})

	asset, err := cld.cloudinaryInstance.Media(location)

	if err != nil {
		return nil, fmt.Errorf("failed to get asset, %v", err)
	}

	asset.String()
	return
}

func (cld *CloudinaryStore) BulkSave(buf chan<- string, listOfRaw [][]byte) {

}

func (cld *CloudinaryStore) Remove(fileName string) {

}
