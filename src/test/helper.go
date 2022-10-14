package test

import (
	"fmt"
	"github.com/hauke96/sigolo"
	"os"
	"path"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"
)

const CacheFolder = "../.test-cache"

func CleanRun(m *testing.M, subFolderName string) {
	Cleanup()
	err := os.MkdirAll(GetCacheFolder(subFolderName), os.ModePerm)
	sigolo.FatalCheck(err)

	m.Run()
}

func Cleanup() {
	err := os.RemoveAll(CacheFolder)
	if err != nil && !os.IsNotExist(err) {
		sigolo.Fatal("Removing %s failed: %s", CacheFolder, err.Error())
	}
}

func GetCacheFolder(subFolderName string) string {
	return path.Join(CacheFolder, subFolderName)
}

func AssertEqual(t *testing.T, expected interface{}, actual interface{}) {
	switch expected.(type) {
	case string:
		expected = strings.ReplaceAll(expected.(string), "\n", "\\n\n")
	}
	switch actual.(type) {
	case string:
		actual = strings.ReplaceAll(actual.(string), "\n", "\\n\n")
	}
	if !reflect.DeepEqual(expected, actual) {
		sigolo.Errorb(1, "Expect to be equal.\nExpected: %+v\n----------\nActual  : %+v", expected, actual)
		t.Fail()
	}
}

func AssertMapEqual(t *testing.T, expected map[string]string, actual map[string]string) {
	if !reflect.DeepEqual(expected, actual) {
		var expectedMapLines []string
		for k, v := range expected {
			s := fmt.Sprintf("> '%s' -> '%s'", k, v)
			s = strings.ReplaceAll(s, "\n", "\\n\n")
			expectedMapLines = append(expectedMapLines, s)
		}
		sort.Strings(expectedMapLines)
		expectedMapString := strings.Join(expectedMapLines, "\n")

		var actualMapLines []string
		for k, v := range actual {
			s := fmt.Sprintf("> '%s' -> '%s'", k, v)
			s = strings.ReplaceAll(s, "\n", "\\n\n")
			actualMapLines = append(actualMapLines, s)
		}
		sort.Strings(actualMapLines)
		actualMapString := strings.Join(actualMapLines, "\n")

		sigolo.Errorb(1, "Expect to be equal.\nExpected:\n%s\n----------\nActual:\n%s", expectedMapString, actualMapString)
		t.Fail()
	}
}

func AssertNil(t *testing.T, value interface{}) {
	if !reflect.DeepEqual(nil, value) {
		sigolo.Errorb(1, "Expect to be 'nil' but was: %+v", value)
		t.Fail()
	}
}

func AssertNotNil(t *testing.T, value interface{}) {
	if nil == value {
		sigolo.Errorb(1, "Expect NOT to be 'nil'", value)
		t.Fail()
	}
}

func AssertError(t *testing.T, expectedMessage string, err error) {
	if expectedMessage != err.Error() {
		sigolo.Errorb(1, "Expected message: %s\nActual error message: %s", expectedMessage, err.Error())
		t.Fail()
	}
}

func AssertEmptyString(t *testing.T, s string) {
	if "" != s {
		sigolo.Errorb(1, "Expected: empty string\nActual  : %s", s)
		t.Fail()
	}
}

func AssertTrue(t *testing.T, b bool) {
	if !b {
		sigolo.Errorb(1, "Expected true but got false")
		t.Fail()
	}
}

func AssertFalse(t *testing.T, b bool) {
	if b {
		sigolo.Errorb(1, "Expected false but got true")
		t.Fail()
	}
}

func AssertMatch(t *testing.T, regexString string, content string) {
	regex := regexp.MustCompile(regexString)
	if !regex.MatchString(content) {
		sigolo.Errorb(1, "Expected to match\nRegex: %s\nContent: %s", regexString, content)
		t.Fail()
	}
}

func AssertNoMatch(t *testing.T, regexString string, content string) {
	regex := regexp.MustCompile(regexString)
	if regex.MatchString(content) {
		sigolo.Errorb(1, "Expected NOT to match\nRegex: %s\nContent: %s", regexString, content)
		t.Fail()
	}
}
