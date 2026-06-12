package storage

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/google/uuid"
)

// LocalStore a FileSystem interface definition over localstorage
type LocalStore struct {
	writes       int
	reads        int
	baseLocation string
	mu           sync.Mutex
}

func NewLocalStore(path string) *LocalStore {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		// MkdirAll creates parent directories as needed and is a no-op if the
		// path already exists, unlike Mkdir which fails when a parent is missing.
		if err := os.MkdirAll(path, 0o755); err != nil {
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

func (l *LocalStore) Save(raw []byte, ext string) (filename string, err error) {
	if len(raw) < 1 {
		return filename, fmt.Errorf("empty file")
	}
	filename = genFileName(ext)
	file, err := NewFile(filepath.Join(l.baseLocation, filename))
	if err != nil {
		return "", err
	}
	defer file.Close()

	if _, err = file.Write(raw); err != nil {
		return "", err
	}

	l.mu.Lock()
	l.writes++
	l.mu.Unlock()
	return filename, nil
}

func (l *LocalStore) Read(fileName string) (file File, err error) {
	file, err = OpenFile(filepath.Join(l.baseLocation, fileName))
	l.mu.Lock()
	l.reads++
	l.mu.Unlock()
	return
}

func (l *LocalStore) BulkSave(buf chan<- FileMeta, listOfFileMeta []FileMeta) {
	var wg sync.WaitGroup
	for _, fd := range listOfFileMeta {
		wg.Add(1)
		go func(raw []byte, title, ext string) {
			defer wg.Done()
			fileName, err := l.Save(raw, ext)
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

func (f *LocalFile) Close() error {
	return f.content.Close()
}

func (f *LocalFile) WriteAt(p []byte, off int64) (n int, err error) {
	return f.content.WriteAt(p, off)
}

func genFileName(ext string) string {
	if ext == "" {
		ext = "mp3"
	}
	return "aud_file_" + uuid.New().String() + "." + ext
}

var cwd, _ = os.Getwd()
var rootDir, _ = filepath.Abs(cwd)
var Store FileSystem = NewLocalStore(filepath.Join(rootDir, "uploads"))
