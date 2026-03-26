package storage

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	s3API "github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// S3Store implements FileSystem over AWS S3 storage.
type S3Store struct {
	session   *session.Session
	bucketId  string
	uploads   int
	downloads int
	mu        sync.Mutex
}

func NewS3Store(bucketId string) *S3Store {
	sess := session.Must(session.NewSession())
	return &S3Store{session: sess, bucketId: bucketId}
}

func (s3 *S3Store) Reads() int {
	return s3.downloads
}

func (s3 *S3Store) Writes() int {
	return s3.uploads
}

func (s3 *S3Store) Save(raw []byte) (filename string, err error) {
	if len(raw) < 1 {
		return "", fmt.Errorf("empty file")
	}
	filename = genFileName()

	freader := bytes.NewReader(raw)

	uploader := s3manager.NewUploader(s3.session)
	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s3.bucketId),
		Key:    aws.String(filename),
		Body:   freader,
	})

	if err != nil {
		return "", fmt.Errorf("failed to upload file, %v", err)
	}

	filename = result.Location

	s3.mu.Lock()
	s3.uploads++
	s3.mu.Unlock()
	return
}

func (s3 *S3Store) Read(location string) (File, error) {
	downloader := s3manager.NewDownloader(s3.session)

	lfile, err := NewFile(os.TempDir() + location)
	if err != nil {
		return nil, fmt.Errorf("failed to create file %q: %w", location, err)
	}

	if _, err = downloader.Download(lfile, &s3API.GetObjectInput{
		Bucket: aws.String(s3.bucketId),
		Key:    aws.String(location),
	}); err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}

	s3.mu.Lock()
	s3.downloads++
	s3.mu.Unlock()

	return lfile, nil
}

func (s3 *S3Store) BulkSave(buf chan<- FileMeta, listOfFileMeta []FileMeta) {
	var wg sync.WaitGroup
	for _, fd := range listOfFileMeta {
		wg.Add(1)
		go func(raw []byte, title string) {
			defer wg.Done()
			fileName, err := s3.Save(raw)
			if err != nil {
				log.Println(err)
				buf <- FileMeta{AudioTitle: title}
				return
			}
			buf <- FileMeta{Name: fileName, Content: raw, AudioTitle: title}
		}(fd.Content, fd.AudioTitle)
	}
	wg.Wait()
	close(buf)
}

func (s3 *S3Store) Remove(fileName string) error {
	s3Client := s3API.New(s3.session)
	_, err := s3Client.DeleteObject(&s3API.DeleteObjectInput{
		Bucket: aws.String(s3.bucketId),
		Key:    aws.String(fileName),
	})
	return err
}
