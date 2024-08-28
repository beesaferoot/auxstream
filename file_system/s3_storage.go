package filesystem

import (
	"bytes"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	s3API "github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"log"
	"sync"
)

// S3Store a FileSystem interface def over s3 storage
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

	// Create an uploader with the session
	uploader := s3manager.NewUploader(s3.session)

	// Upload the file to S3.
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

func (s3 *S3Store) Read(location string) (file File, err error) {
	// Create a downloader with the session=
	downloader := s3manager.NewDownloader(s3.session)
	// Create a file to write the S3 Object contents to.
	lfile, err := NewFile(location)
	file = lfile

	if err != nil {
		return nil, fmt.Errorf("failed to create file %q, %v", location, err)
	}

	_, err = downloader.Download(lfile, &s3API.GetObjectInput{
		Bucket: aws.String(s3.bucketId),
		Key:    aws.String(location),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to download file, %v", err)
	}

	s3.mu.Lock()
	s3.downloads++
	s3.mu.Unlock()

	return
}

func (s3 *S3Store) BulkSave(buf chan<- string, listOfRaw [][]byte) {
	var wg sync.WaitGroup
	for _, raw := range listOfRaw {
		wg.Add(1)
		go func(raw []byte) {
			defer wg.Done()
			fileName, err := s3.Save(raw)
			if err != nil {
				log.Println(err)
				buf <- ""
				return
			}
			buf <- fileName
		}(raw)
	}
	wg.Wait()
	close(buf)
}

func (s3 *S3Store) Remove(fileName string) error {
	deleteObjectInput := &s3API.DeleteObjectInput{
		Bucket: aws.String(s3.bucketId),
		Key:    aws.String(fileName),
	}

	s3Client := s3API.New(s3.session)

	// Delete the object
	_, err := s3Client.DeleteObject(deleteObjectInput)
	if err != nil {
		log.Println("Error deleting object:", err)
		return err
	}

	log.Printf("Object '%s' deleted from bucket '%s'\n", fileName, s3.bucketId)
	return nil
}
