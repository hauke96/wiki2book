package test

import (
	"reflect"
	"regexp"
	"testing"
)

func AssertEqual(t *testing.T, expected interface{}, actual interface{}) {
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expect to be equal.\nExpected: %+v\nActual: %+v", expected, actual)
		t.Fail()
	}
}

func AssertError(t *testing.T, expectedMessage string, err error) {
	if expectedMessage != err.Error() {
		t.Errorf("Expected message: %s\nActual error message: %s", expectedMessage, err.Error())
		t.Fail()
	}
}

func AssertEmptyString(t *testing.T, s string) {
	if "" != s {
		t.Errorf("Expected: empty string\nActual: %s", s)
		t.Fail()
	}
}

func AssertTrue(t *testing.T, b bool) {
	if !b {
		t.Error("Expected true but got false")
		t.Fail()
	}
}

func AssertFalse(t *testing.T, b bool) {
	if b {
		t.Error("Expected false but got true")
		t.Fail()
	}
}

func AssertMatch(t *testing.T, regexString string, content string) {
	regex := regexp.MustCompile(regexString)
	if !regex.MatchString(content) {
		t.Errorf("Expected to match\nRegex: %s\nContent: %s", regexString, content)
		t.Fail()
	}
}

func AssertNoMatch(t *testing.T, regexString string, content string) {
	regex := regexp.MustCompile(regexString)
	if regex.MatchString(content) {
		t.Errorf("Expected NOT to match\nRegex: %s\nContent: %s", regexString, content)
		t.Fail()
	}
}

