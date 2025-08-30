package util

import (
	"os"
)

type MockFilesystem struct {
	ExistsFunc          func(path string) bool
	GetSizeInBytesFunc  func(path string) (int64, error)
	RenameFunc          func(oldPath string, newPath string) error
	RemoveFunc          func(name string) error
	MkdirAllFunc        func(path string) error
	CreateTempFunc      func(dir, pattern string) (*os.File, error)
	DirSizeInBytesFunc  func(path string) (error, int64)
	FindLargestFileFunc func(path string) (error, float64, string)
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

func (m *MockFilesystem) CreateTemp(dir, filenamePattern string) (*os.File, error) {
	return m.CreateTempFunc(dir, filenamePattern)
}

func (m *MockFilesystem) Remove(path string) error {
	return m.RemoveFunc(path)
}

func (m *MockFilesystem) DirSizeInBytes(path string) (error, int64) {
	return m.DirSizeInBytesFunc(path)
}

func (m *MockFilesystem) FindLargestFile(path string) (error, float64, string) {
	return m.FindLargestFileFunc(path)
}
