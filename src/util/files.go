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
		err = errors.Wrapf(err, "Unable to make file path '%s' relative", path)
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
		err = errors.Wrapf(err, "Unable to make file path '%s' absolute", path)
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

// SanitizeFilename URL encodes characters, that are problematic for file paths. This function assumes, that filenames
// do not contain non-printable characters (e.g. ASCII 0-31) or reserved words like "COM1" on Windows or "." on Linux.
// Only usually invalid printable characters are encoded, all other characters (even special characters) stay unchanged.
func SanitizeFilename(filename string) string {
	/*
		Not allowed in Windows filesystems:
		  < (less than)
		  > (greater than)
		  : (colon)
		  " (double quote)
		  / (forward slash)
		  \ (backslash)
		  | (vertical bar or pipe)
		  ? (question mark)
		  * (asterisk)

		Not allowed on Linux and OS/X filesystems:
		  / (forward slash)
	*/

	filename = strings.ReplaceAll(filename, "%", "%25")
	filename = strings.ReplaceAll(filename, "<", "%3C")
	filename = strings.ReplaceAll(filename, ">", "%3E")
	filename = strings.ReplaceAll(filename, ":", "%3A")
	filename = strings.ReplaceAll(filename, "\"", "%22")
	filename = strings.ReplaceAll(filename, "/", "%2F")
	filename = strings.ReplaceAll(filename, "\\", "%5C")
	filename = strings.ReplaceAll(filename, "|", "%7C")
	filename = strings.ReplaceAll(filename, "?", "%3F")
	filename = strings.ReplaceAll(filename, "*", "%2A")

	return filename
}

// IsSanitized URL encodes characters, that are problematic for file paths. This function assumes, that filenames
// do not contain non-printable characters (e.g. ASCII 0-31) or reserved words like "COM1" on Windows or "." on Linux.
// Only usually invalid printable characters are encoded, all other characters (even special characters) stay unchanged.
func IsSanitized(path string) bool {
	/*
		Not allowed in Windows filesystems:
		  < (less than)
		  > (greater than)
		  : (colon)
		  " (double quote)
		  / (forward slash)
		  \ (backslash)
		  | (vertical bar or pipe)
		  ? (question mark)
		  * (asterisk)

		Not allowed on Linux and OS/X filesystems:
		  / (forward slash)
	*/

	parts := strings.SplitN(path, string(filepath.Separator), -1)
	for _, part := range parts {
		// The escape sequences are fine, but no other "%" should exist, since it indicates an un-sanitized "%" character.
		part = strings.ReplaceAll(part, "%25", "")
		part = strings.ReplaceAll(part, "%3C", "")
		part = strings.ReplaceAll(part, "%3E", "")
		part = strings.ReplaceAll(part, "%3A", "")
		part = strings.ReplaceAll(part, "%22", "")
		part = strings.ReplaceAll(part, "%2F", "")
		part = strings.ReplaceAll(part, "%5C", "")
		part = strings.ReplaceAll(part, "%7C", "")
		part = strings.ReplaceAll(part, "%3F", "")
		part = strings.ReplaceAll(part, "%2A", "")

		if strings.Contains(part, "%") ||
			strings.Contains(part, "<") ||
			strings.Contains(part, ">") ||
			strings.Contains(part, ":") ||
			strings.Contains(part, "\"") ||
			strings.Contains(part, "/") ||
			strings.Contains(part, "\\") ||
			strings.Contains(part, "|") ||
			strings.Contains(part, "?") ||
			strings.Contains(part, "*") {
			return false
		}
	}

	return true
}

func RequireFilePathIsSanitized(path string) {
	Requiref(IsSanitized(path), "File path or name '%s' mus be sanitized.", path)
}

type FileLike interface {
	Name() string
	Write(p []byte) (n int, err error)
	Stat() (os.FileInfo, error)
	Close() error
}

type Filesystem interface {
	GetSizeInBytes(path string) (int64, error)
	Rename(oldPath string, newPath string) error
	Remove(name string) error
	Create(name string) (FileLike, error)
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

func (o *OsFilesystem) GetSizeInBytes(path string) (int64, error) {
	RequireFilePathIsSanitized(path)

	stat, err := os.Stat(path)
	if err != nil {
		return math.MinInt64, err
	}
	return stat.Size(), nil
}

func (o *OsFilesystem) Rename(oldPath string, newPath string) error {
	RequireFilePathIsSanitized(oldPath)
	RequireFilePathIsSanitized(newPath)

	return os.Rename(oldPath, newPath)
}

func (o *OsFilesystem) MkdirAll(path string) error {
	RequireFilePathIsSanitized(path)

	return os.MkdirAll(path, os.ModePerm)
}

func (o *OsFilesystem) CreateTemp(dir, filenamePattern string) (FileLike, error) {
	RequireFilePathIsSanitized(dir)
	RequireFilePathIsSanitized(filenamePattern)

	return os.CreateTemp(dir, filenamePattern)
}

func (o *OsFilesystem) Remove(path string) error {
	RequireFilePathIsSanitized(path)

	return os.Remove(path)
}

func (o *OsFilesystem) Create(path string) (FileLike, error) {
	RequireFilePathIsSanitized(path)

	return os.Create(path)
}

func (o *OsFilesystem) DirSizeInBytes(path string) (error, int64) {
	RequireFilePathIsSanitized(path)

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
	RequireFilePathIsSanitized(path)
	RequireFilePathIsSanitized(exceptDir)

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
	RequireFilePathIsSanitized(path)
	RequireFilePathIsSanitized(exceptDir)

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
	RequireFilePathIsSanitized(name)

	return os.ReadFile(name)
}

func (o *OsFilesystem) Stat(name string) (os.FileInfo, error) {
	RequireFilePathIsSanitized(name)

	return os.Stat(name)
}

func (o *OsFilesystem) Chtimes(name string, atime time.Time, mtime time.Time) error {
	RequireFilePathIsSanitized(name)

	return os.Chtimes(name, atime, mtime)
}
