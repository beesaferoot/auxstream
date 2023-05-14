package filesystem

import (
	"os"
	"sync"
)

/* FileSystem interface definition over localstorage */
type LocalStore struct {
	writes       int
	reads        int
	baseLocation string
}

func NewStore(path string) *LocalStore {
	return &LocalStore{baseLocation: path, reads: 0, writes: 0}
}

func (l *LocalStore) Reads() int {
	return l.reads
}

func (l *LocalStore) Writes() int {
	return l.writes
}

func (l *LocalStore) Save(file_name string, raw []byte) (err error) {
	file := NewFile(l.baseLocation + "/" + file_name)
	_, err = file.Write(raw)
	l.writes++
	return
}

func (l *LocalStore) Read(file_name string) (file File, err error) {
	file, err = OpenFile(l.baseLocation + "/" + file_name)
	l.reads++
	return
}

func (l *LocalStore) BulkSave(list_of_file_names []string, list_of_raw [][]byte) {
	var wg sync.WaitGroup

	for i, raw := range list_of_raw {
		wg.Add(1)
		go func(file_name string, raw []byte) {
			defer wg.Done()
			l.Save(file_name, raw)
		}(list_of_file_names[i], raw)
	}
	wg.Wait()
}

func (l *LocalStore) Remove(file_name string) error {
	return os.Remove(l.baseLocation + "/" + file_name)
}

/* File interface definition over localstorage */
type LocalFile struct {
	filePath string
	content  *os.File
}

func NewFile(filePath string) *LocalFile {
	content, _ := os.Create(filePath)
	return &LocalFile{filePath: filePath, content: content}
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
