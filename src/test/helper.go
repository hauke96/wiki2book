package test

import (
	"fmt"
	"github.com/hauke96/sigolo/v2"
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

const tempDirName = ".tmp/"
const cacheFolder = "../.test-cache"

func CleanRun(m *testing.M) {
	Cleanup()
	err := os.MkdirAll(GetCacheFolder(), os.ModePerm)
	sigolo.FatalCheck(err)

	err = os.RemoveAll(tempDirName)
	sigolo.FatalCheck(err)
	err = os.MkdirAll(tempDirName, os.ModePerm)
	sigolo.FatalCheck(err)

	m.Run()
}

func Cleanup() {
	err := os.RemoveAll(cacheFolder)
	if err != nil && !os.IsNotExist(err) {
		sigolo.Fatalf("Removing %s failed: %s", cacheFolder, err.Error())
	}
}

func GetCacheFolder() string {
	return cacheFolder
}

func AssertEqual(t *testing.T, expected interface{}, actual interface{}) {
	expectedIsString := false
	actualIsString := false

	switch expected.(type) {
	case string:
		expectedIsString = true
	}
	switch actual.(type) {
	case string:
		actualIsString = true
	}

	if !reflect.DeepEqual(expected, actual) {
		if expectedIsString && actualIsString {
			assertEqualStrings(t, expected.(string), actual.(string))
		} else {
			sigolo.Errorb(1, "Expect to be equal.\nExpected: %+v\n----------\nActual  : %+v", expected, actual)
			t.Fail()
		}
	}
}

func assertEqualStrings(t *testing.T, expected string, actual string) {
	expected = strings.ReplaceAll(expected, "\n", "\\n\n")

	actual = strings.ReplaceAll(actual, "\n", "\\n\n")

	expectedLines := strings.Split(expected, "\n")
	actualLines := strings.Split(actual, "\n")

	sigolo.Errorb(2, "Expect to be equal.\n|   | %-50s | %-50s |", "Expected", "Actual")
	fmt.Printf("|%s|\n", strings.Repeat("-", 109))

	for i, expectedLine := range expectedLines {
		actualLine := ""
		if len(actualLines) > i {
			actualLine = actualLines[i]
		}

		changeMark := " "
		if actualLine != expectedLine {
			changeMark = "*"
		}

		fmt.Printf("| %s | %-50s | %-50s |\n", changeMark, "\""+expectedLine+"\"", "\""+actualLine+"\"")
	}

	if len(actualLines) > len(expectedLines) {
		for i := len(expectedLines); i < len(actualLines); i++ {
			actualLine := actualLines[i]
			fmt.Printf("| * | %-50s | %-50s |\n", "", "\""+actualLine+"\"")
		}
	}

	t.Fail()
}

func AssertMapEqual[K comparable, V comparable](t *testing.T, expected map[K]V, actual map[K]V) {
	if !reflect.DeepEqual(expected, actual) {
		expectedMapString := compareMaps(expected, "E", actual, "A")

		var notExpectedValues []string
		for actualK, actualV := range actual {
			valueInExpectedMap, expectedMapHasKey := expected[actualK]
			valueIsNotInExpectedMap := !expectedMapHasKey || !reflect.DeepEqual(valueInExpectedMap, actualV)

			vString := fmt.Sprintf("%v", actualV)
			if valueIsNotInExpectedMap {
				vString = strings.ReplaceAll(vString, "\n", "\n  ")
				s := fmt.Sprintf("  '%v' -> '%s'", actualK, vString)
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
func compareMaps[K comparable, V comparable](values map[K]V, valuePrefix string, otherValues map[K]V, otherValuePrefix string) string {
	var expectedMapLines []string
	for k, v := range values {
		linePrefix, valueIsNotInOtherMap := getLinePrefix(otherValues, k, v, valuePrefix)

		vString := fmt.Sprintf("%v", v)
		vString = strings.ReplaceAll(vString, "\n", "\n  ")
		s := fmt.Sprintf("%s '%v' -> '%s'", linePrefix, k, vString)
		s = strings.ReplaceAll(s, "\n", "\\n\n")
		expectedMapLines = append(expectedMapLines, s)

		if valueIsNotInOtherMap {
			v = otherValues[k]
			vString = fmt.Sprintf("%v", v)
			vString = strings.ReplaceAll(vString, "\n", "\n  ")
			s := fmt.Sprintf("%s '%v' -> '%s'", otherValuePrefix, k, vString)
			s = strings.ReplaceAll(s, "\n", "\\n\n")
			expectedMapLines = append(expectedMapLines, s)
		}
	}
	expectedMapString := strings.Join(expectedMapLines, "\n")
	return expectedMapString
}

// getLinePrefix returns the prefix for map comparison lines. It can be used to mark lines not contains in the
// expected/actual result map.
func getLinePrefix[K comparable, V comparable](otherMap map[K]V, key K, expectedValue V, prefixIfNotInOtherMap string) (string, bool) {
	valueInOtherMap, otherMapHasKey := otherMap[key]
	valueIsNotInOtherMap := !otherMapHasKey || !reflect.DeepEqual(valueInOtherMap, expectedValue)

	if valueIsNotInOtherMap {
		return prefixIfNotInOtherMap, valueIsNotInOtherMap
	}
	return " ", valueIsNotInOtherMap
}

func AssertNil(t *testing.T, value any) {
	if value != nil && !reflect.ValueOf(value).IsNil() {
		sigolo.Errorb(1, "Expect to be 'nil' but was: %#v", value)
		t.Fail()
	}
}

func AssertNotNil(t *testing.T, value any) {
	if value == nil || reflect.ValueOf(value).IsNil() {
		sigolo.Errorb(1, "Expect NOT to be 'nil' but was: %#v", value)
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
