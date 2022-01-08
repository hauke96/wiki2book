package test

import "testing"

func AssertEqualString(t *testing.T, expected string, actual string) {
	if expected != actual {
		t.Errorf("Expected: %s\nActual: %s", expected, actual)
		t.Fail()
	}
}

func AssertEmptyString(t *testing.T, s string) {
	if "" != s {
		t.Errorf("Expected: empty string\nActual: %s", s)
		t.Fail()
	}
}
