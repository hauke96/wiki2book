package test

import "testing"

func AssertEqual(t *testing.T, expected string, actual string) {
	if expected != actual {
		t.Errorf("Expected: %s\nActual: %s", expected, actual)
		t.Fail()
	}
}
