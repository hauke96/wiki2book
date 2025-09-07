package util

import (
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hauke96/sigolo/v2"
	"github.com/pkg/errors"
)

const (
	FileEndingSvg  = ".svg"
	FileEndingPng  = ".png"
	FileEndingPdf  = ".pdf"
	FileEndingWebp = ".webp"
)

func ToRelativePaths(paths ...string) ([]string, error) {
	var result = make([]string, len(paths))

	for i, path := range paths {
		relativePath, err := ToRelativePath(path)
		if err != nil {
			return nil, err
		}
		result[i] = relativePath
	}

	return result, nil
}

func ToRelativePath(path string) (string, error) {
	currentDir, err := os.Getwd()
	sigolo.FatalCheck(err)

	if path != "" {
		path, err = filepath.Rel(currentDir, path)
		err = errors.Wrapf(err, "Unable to make file path %s relative", path)
	}

	return path, err
}

func ToRelativePathWithBasedir(basedir string, path string) (string, error) {
	var err error
	if path != "" {
		path, err = filepath.Rel(basedir, path)
		err = errors.Wrapf(err, "Unable to make file path '%s' relative", path)
	}

	return path, nil
}

func ToAbsolutePaths(paths ...string) ([]string, error) {
	var result = make([]string, len(paths))

	for i, path := range paths {
		absolutePath, err := ToAbsolutePath(path)
		if err != nil {
			return nil, err
		}
		result[i] = absolutePath
	}

	return result, nil
}

func ToAbsolutePath(path string) (string, error) {
	var err error
	if path != "" && !filepath.IsAbs(path) {
		path, err = filepath.Abs(path)
		err = errors.Wrapf(err, "Unable to make file path %s absolute", path)
	}
	return path, err
}

func ToAbsolutePathWithBasedir(basedir string, path string) (string, error) {
	var err error
	if path != "" && !filepath.IsAbs(path) {
		path, err = filepath.Abs(filepath.Join(basedir, path))
		err = errors.Wrapf(err, "Unable to make file path '%s' absolute", filepath.Join(basedir, path))
	}
	return path, err
}

func AssertPathExists(path string) {
	if !PathExists(path) {
		sigolo.FatalCheck(errors.Errorf("Path '%s' does not exist", path))
	}
}

func PathExists(path string) bool {
	if _, err := os.Stat(path); strings.TrimSpace(path) != "" && err != nil {
		return false
	}
	return true
}

func EnsureDirectory(path string) {
	if !PathExists(path) {
		sigolo.Debugf("Create directory '%s'", path)
		err := os.MkdirAll(path, os.ModePerm)
		sigolo.FatalCheck(errors.Wrapf(err, "Error creating '%s' directory", path))
	} else {
		sigolo.Debugf("Directory '%s' already exists", path)
	}
}

// GetPngPathForPdf converts the given path of a PDF file into a PNG file.
func GetPngPathForPdf(path string) string {
	Requiref(filepath.Ext(strings.ToLower(path)) == FileEndingPdf, "Filepath must lead to a PDF file but was '%s'", path)
	return path + FileEndingPng
}

// GetPngPathForSvg converts the given path of a PDF file into a PNG file.
func GetPngPathForSvg(path string) string {
	Requiref(filepath.Ext(strings.ToLower(path)) == FileEndingSvg, "Filepath must lead to a SVG file but was '%s'", path)
	return path + FileEndingPng
}

// GetPngPathForFile converts the given path into a PNG file path.
func GetPngPathForFile(path string) string {
	return path + FileEndingPng
}

func ToMB(size int64) float64 {
	if size == math.MinInt64 {
		return math.NaN()
	}
	return float64(size) / 1024.0 / 1024.0
}

type FileLike interface {
	Name() string
	Write(p []byte) (n int, err error)
	Stat() (os.FileInfo, error)
}

type Filesystem interface {
	Exists(path string) bool
	GetSizeInBytes(path string) (int64, error)
	Rename(oldPath string, newPath string) error
	Remove(name string) error
	MkdirAll(path string) error
	CreateTemp(dir, pattern string) (FileLike, error)
	DirSizeInBytes(path string) (error, int64)
	FindLargestFile(path string, exceptDir string) (error, int64, string)
	FindLruFile(path string, exceptDir string) (error, int64, string)
	ReadFile(name string) ([]byte, error)
	Stat(name string) (os.FileInfo, error)
	Chtimes(name string, atime time.Time, mtime time.Time) error
}

var CurrentFilesystem Filesystem = &OsFilesystem{}

type OsFilesystem struct {
}

func (o *OsFilesystem) Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (o *OsFilesystem) GetSizeInBytes(path string) (int64, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return math.MinInt64, err
	}
	return stat.Size(), nil
}

func (o *OsFilesystem) Rename(oldPath string, newPath string) error {
	return os.Rename(oldPath, newPath)
}

func (o *OsFilesystem) MkdirAll(path string) error {
	return os.MkdirAll(path, os.ModePerm)
}

func (o *OsFilesystem) CreateTemp(dir, filenamePattern string) (FileLike, error) {
	return os.CreateTemp(dir, filenamePattern)
}

func (o *OsFilesystem) Remove(path string) error {
	return os.Remove(path)
}

func (o *OsFilesystem) DirSizeInBytes(path string) (error, int64) {
	var dirSizeBytes int64 = 0

	readSize := func(path string, file os.FileInfo, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				// It might happen, that files are deleted during file walk (concurrency etc.)
				return nil
			}
			return err
		}
		if !file.IsDir() {
			dirSizeBytes += file.Size()
		}
		return nil
	}

	err := filepath.Walk(path, readSize)
	if err != nil {
		return err, math.MinInt64
	}

	return nil, dirSizeBytes
}

func (o *OsFilesystem) FindLargestFile(path string, exceptDir string) (error, int64, string) {
	var currentLargestFile os.FileInfo
	var currentLargestFilePath string

	readSize := func(path string, file os.FileInfo, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				// It might happen, that files are deleted during file walk (concurrency etc.)
				return nil
			}
			return err
		}
		if !file.IsDir() {
			if currentLargestFile == nil || file.Size() > currentLargestFile.Size() {
				currentLargestFile = file
				currentLargestFilePath = path
			}
		} else if file.Name() == exceptDir {
			// The directory for temporary files might be inside the cache folder. This is fine, but we don't want to
			// count in temporary files then.
			return filepath.SkipDir
		}
		return nil
	}

	err := filepath.Walk(path, readSize)
	if err != nil {
		return err, math.MinInt64, ""
	}

	return nil, currentLargestFile.Size(), currentLargestFilePath
}

func (o *OsFilesystem) FindLruFile(path string, exceptDir string) (error, int64, string) {
	var currentLruFilePath string
	var currentLruFile os.FileInfo

	readModTime := func(path string, file os.FileInfo, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				// It might happen, that files are deleted during file walk (concurrency etc.)
				return nil
			}
			return err
		}
		if !file.IsDir() {
			if currentLruFile == nil || file.ModTime().Before(currentLruFile.ModTime()) {
				currentLruFile = file
				currentLruFilePath = path
			}
		} else if file.Name() == exceptDir {
			// The directory for temporary files might be inside the cache folder. This is fine, but we don't want to
			// count in temporary files then.
			return filepath.SkipDir
		}
		return nil
	}

	err := filepath.Walk(path, readModTime)
	if err != nil {
		return err, math.MinInt64, ""
	}

	if currentLruFile == nil {
		// No file was found
		return nil, 0, ""
	}

	return nil, currentLruFile.Size(), currentLruFilePath
}

func (o *OsFilesystem) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}

func (o *OsFilesystem) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

func (o *OsFilesystem) Chtimes(name string, atime time.Time, mtime time.Time) error {
	return os.Chtimes(name, atime, mtime)
}
