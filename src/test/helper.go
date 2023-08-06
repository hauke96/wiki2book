package test

import (
	"fmt"
	"github.com/hauke96/sigolo"
	"os"
	"path"
	"reflect"
	"regexp"
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
		expectedMapString := compareMaps(expected, "E", actual, "A")

		var notExpectedValues []string
		for k, v := range actual {
			valueInOtherMap, otherMapHasKey := expected[k]
			valueIsNotInExpectedMap := !otherMapHasKey || valueInOtherMap != expected[k]

			if valueIsNotInExpectedMap {
				v = strings.ReplaceAll(v, "\n", "\n  ")
				s := fmt.Sprintf("  '%s' -> '%s'", k, v)
				s = strings.ReplaceAll(s, "\n", "\\n\n")
				notExpectedValues = append(notExpectedValues, s)
			}
		}
		notExpectedValuesString := strings.Join(notExpectedValues, "\n")

		sigolo.Errorb(1, `Expect to be equal.

Prefix meanings:
  A: Actual value
  E: Expected value

Expected values:
%s

Values not expected but still found:
%s`, expectedMapString, notExpectedValuesString)
		t.Fail()
	}
}

// compareMaps lists all rows of the values-map and marks rows with the valuePrefix if they are not or in different form
// in the other map
func compareMaps(values map[string]string, valuePrefix string, otherValues map[string]string, otherValuePrefix string) string {
	var expectedMapLines []string
	for k, v := range values {
		linePrefix, valueIsNotInOtherMap := getLinePrefix(otherValues, k, v, valuePrefix)

		v = strings.ReplaceAll(v, "\n", "\n  ")
		s := fmt.Sprintf("%s '%s' -> '%s'", linePrefix, k, v)
		s = strings.ReplaceAll(s, "\n", "\\n\n")
		expectedMapLines = append(expectedMapLines, s)

		if valueIsNotInOtherMap {
			v = otherValues[k]
			v = strings.ReplaceAll(v, "\n", "\n  ")
			s := fmt.Sprintf("%s '%s' -> '%s'", otherValuePrefix, k, v)
			s = strings.ReplaceAll(s, "\n", "\\n\n")
			expectedMapLines = append(expectedMapLines, s)
		}
	}
	expectedMapString := strings.Join(expectedMapLines, "\n")
	return expectedMapString
}

// getLinePrefix returns the prefix for map comparison lines. It can be used to mark lines not contains in the
// expected/actual result map.
func getLinePrefix(otherMap map[string]string, key string, expectedValue string, prefixIfNotInOtherMap string) (string, bool) {
	valueInOtherMap, otherMapHasKey := otherMap[key]
	valueIsNotInOtherMap := !otherMapHasKey || valueInOtherMap != expectedValue

	if valueIsNotInOtherMap {
		return prefixIfNotInOtherMap, valueIsNotInOtherMap
	}
	return " ", valueIsNotInOtherMap
}

func AssertNil(t *testing.T, value interface{}) {
	if !reflect.DeepEqual(nil, value) {
		sigolo.Errorb(1, "Expect to be 'nil' but was: %+v", value)
		t.Fail()
	}
}

func AssertNotNil(t *testing.T, value interface{}) {
	if nil == value {
		sigolo.Errorb(1, "Expect NOT to be 'nil' but was: %+v", value)
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
