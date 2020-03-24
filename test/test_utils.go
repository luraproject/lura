package test

import "testing"

func AssertEqual(t *testing.T, expected, actual, message string) {
	if actual != expected {
		t.Errorf("%s. Expected: %s, actual %s.", message, expected, actual)
	}
}
