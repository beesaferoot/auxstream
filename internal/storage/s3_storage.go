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

// S3Store is a FileSystem backed by an AWS S3 bucket. Save returns the object's S3
// URL (its Location) as the identifier; Read and Remove instead key off the object
// key, so a URL from Save is not directly reusable as the argument to either.
type S3Store struct {
	session   *session.Session
	bucketId  string
	uploads   int
	downloads int
	mu        sync.Mutex
}

// NewS3Store targets bucketId, drawing region and credentials from the AWS default
// chain (environment, shared config, instance role). Construction is fatal if no
// session can be established.
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

func (s3 *S3Store) Save(raw []byte, ext string) (filename string, err error) {
	if len(raw) < 1 {
		return "", fmt.Errorf("empty file")
	}
	filename = genFileName(ext)

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

	// The downloader needs a WriterAt, so stage the object in a local temp file and
	// hand that back as the File; the caller owns closing (and cleaning up) it.
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
		go func(raw []byte, title, ext string) {
			defer wg.Done()
			fileName, err := s3.Save(raw, ext)
			if err != nil {
				log.Println(err)
				buf <- FileMeta{AudioTitle: title}
				return
			}
			buf <- FileMeta{Name: fileName, Content: raw, AudioTitle: title, Ext: ext}
		}(fd.Content, fd.AudioTitle, fd.Ext)
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
