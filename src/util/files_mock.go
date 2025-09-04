package util

import (
	"os"
	"time"
)

type MockFileInfo struct {
	name  string
	size  int64
	mode  os.FileMode
	mod   time.Time
	isDir bool
}

func NewMockFileInfo(name string) os.FileInfo {
	return &MockFileInfo{
		name:  name,
		size:  0,
		mode:  0666,
		mod:   time.Time{},
		isDir: false,
	}
}

func NewMockFileInfoWithTime(name string, time time.Time) os.FileInfo {
	return &MockFileInfo{
		name:  name,
		size:  0,
		mode:  0666,
		mod:   time,
		isDir: false,
	}
}

func (f MockFileInfo) Name() string       { return f.name }
func (f MockFileInfo) Size() int64        { return f.size }
func (f MockFileInfo) Mode() os.FileMode  { return f.mode }
func (f MockFileInfo) ModTime() time.Time { return f.mod }
func (f MockFileInfo) IsDir() bool        { return f.isDir }
func (f MockFileInfo) Sys() interface{}   { return nil }

type MockFile struct {
	NameFunc  func() string
	WriteFunc func(p []byte) (n int, err error)
	StatFunc  func() os.FileInfo

	WrittenBytes []byte
}

func NewMockFile(name string) *MockFile {
	return &MockFile{
		NameFunc: func() string {
			return name
		},
		WriteFunc: func(p []byte) (n int, err error) { return len(p), nil },
	}
}

func (m *MockFile) Name() string {
	return m.NameFunc()
}

func (m *MockFile) Write(p []byte) (n int, err error) {
	m.WrittenBytes = append(m.WrittenBytes, p...)
	return m.WriteFunc(p)
}

func (m *MockFile) Stat() (os.FileInfo, error) {
	return NewMockFileInfo(m.Name()), nil
}

type MockFilesystem struct {
	ExistsFunc          func(path string) bool
	GetSizeInBytesFunc  func(path string) (int64, error)
	RenameFunc          func(oldPath string, newPath string) error
	RemoveFunc          func(name string) error
	MkdirAllFunc        func(path string) error
	CreateTempFunc      func(dir, pattern string) (FileLike, error)
	DirSizeInBytesFunc  func(path string) (error, int64)
	FindLargestFileFunc func(path string, exceptDir string) (error, int64, string)
	FindLruFileFunc     func(path string, exceptDir string) (error, int64, string)
	ReadFileFunc        func(name string) ([]byte, error)
	StatFunc            func(name string) (os.FileInfo, error)
	ChtimesFunc         func(name string, atime time.Time, mtime time.Time) error
}

func NewDefaultMockFilesystem() *MockFilesystem {
	return &MockFilesystem{
		ExistsFunc:          func(path string) bool { return false },
		GetSizeInBytesFunc:  func(path string) (int64, error) { return -1, nil },
		RenameFunc:          func(oldPath string, newPath string) error { return nil },
		RemoveFunc:          func(name string) error { return nil },
		MkdirAllFunc:        func(path string) error { return nil },
		CreateTempFunc:      func(dir, pattern string) (FileLike, error) { return NewMockFile(pattern), nil },
		DirSizeInBytesFunc:  func(path string) (error, int64) { return nil, -1 },
		FindLargestFileFunc: func(path string, exceptDir string) (error, int64, string) { return nil, -1, "" },
		FindLruFileFunc:     func(path string, exceptDir string) (error, int64, string) { return nil, -1, "" },
		ReadFileFunc:        func(name string) ([]byte, error) { return []byte{}, nil },
		StatFunc:            func(name string) (os.FileInfo, error) { return NewMockFileInfo("__mock-file-info__"), nil },
		ChtimesFunc:         func(name string, atime time.Time, mtime time.Time) error { return nil },
	}
}

func (m *MockFilesystem) Exists(path string) bool {
	return m.ExistsFunc(path)
}

func (m *MockFilesystem) GetSizeInBytes(path string) (int64, error) {
	return m.GetSizeInBytesFunc(path)
}

func (m *MockFilesystem) Rename(oldPath string, newPath string) error {
	return m.RenameFunc(oldPath, newPath)
}

func (m *MockFilesystem) MkdirAll(path string) error {
	return m.MkdirAllFunc(path)
}

func (m *MockFilesystem) CreateTemp(dir, filenamePattern string) (FileLike, error) {
	return m.CreateTempFunc(dir, filenamePattern)
}

func (m *MockFilesystem) Remove(path string) error {
	return m.RemoveFunc(path)
}

func (m *MockFilesystem) DirSizeInBytes(path string) (error, int64) {
	return m.DirSizeInBytesFunc(path)
}

func (m *MockFilesystem) FindLargestFile(path string, exceptDir string) (error, int64, string) {
	return m.FindLargestFileFunc(path, exceptDir)
}

func (m *MockFilesystem) FindLruFile(path string, exceptDir string) (error, int64, string) {
	return m.FindLruFileFunc(path, exceptDir)
}

func (m *MockFilesystem) ReadFile(name string) ([]byte, error) {
	return m.ReadFileFunc(name)
}

func (m *MockFilesystem) Stat(name string) (os.FileInfo, error) {
	return m.StatFunc(name)
}

func (m *MockFilesystem) Chtimes(name string, atime time.Time, mtime time.Time) error {
	return m.ChtimesFunc(name, atime, mtime)
}
