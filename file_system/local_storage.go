package filesystem

import (
	"fmt"
	"github.com/google/uuid"
	"log"
	"os"
	"path/filepath"
	"sync"
)

// LocalStore a FileSystem interface definition over localstorage
type LocalStore struct {
	writes       int
	reads        int
	baseLocation string
}

func NewStore(path string) *LocalStore {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		// Create the directory
		err := os.Mkdir(path, 0744)
		if err != nil {
			log.Fatalln(err.Error())
		}
	}
	return &LocalStore{baseLocation: path, reads: 0, writes: 0}
}

func (l *LocalStore) Reads() int {
	return l.reads
}

func (l *LocalStore) Writes() int {
	return l.writes
}

func (l *LocalStore) Save(raw []byte) (filename string, err error) {
	if len(raw) < 1 {
		return filename, fmt.Errorf("empty file")
	}
	filename = genFileName()
	file, err := NewFile(filepath.Join(l.baseLocation, filename))
	if err != nil {
		return "", err
	}
	_, err = file.Write(raw)
	l.writes++
	return
}

func (l *LocalStore) Read(fileName string) (file File, err error) {
	file, err = OpenFile(filepath.Join(l.baseLocation, fileName))
	l.reads++
	return
}

func (l *LocalStore) BulkSave(buf chan<- string, listOfRaw [][]byte) {
	var wg sync.WaitGroup
	for _, raw := range listOfRaw {
		wg.Add(1)
		go func(raw []byte) {
			defer wg.Done()
			fileName, err := l.Save(raw)
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

func (l *LocalStore) Remove(fileName string) error {
	return os.Remove(filepath.Join(l.baseLocation, fileName))
}

// LocalFile  File interface definition over localstorage
type LocalFile struct {
	filePath string
	content  *os.File
}

func NewFile(filePath string) (file *LocalFile, err error) {
	content, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}
	file = &LocalFile{filePath: filePath, content: content}
	return file, nil
}

func OpenFile(filePath string) (*LocalFile, error) {
	content, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	return &LocalFile{filePath: filePath, content: content}, nil
}

func (f *LocalFile) Name() string {
	return f.content.Name()
}

func (f *LocalFile) Size() int64 {
	info, _ := f.content.Stat()
	return info.Size()
}

func (f *LocalFile) Read(p []byte) (n int, err error) {
	return f.content.Read(p)
}

func (f *LocalFile) Write(p []byte) (n int, err error) {
	return f.content.Write(p)
}

func genFileName() string {
	return "aud_file_" + uuid.New().String() + ".mp3"
}

var cwd, _ = os.Getwd()
var rootDir, _ = filepath.Abs(cwd)
var Store = NewStore(filepath.Join(rootDir, "uploads"))
