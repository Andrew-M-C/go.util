package gotest

import "testing"

func Test_IsInGoTest(t *testing.T) {
	if !IsInGoTest() {
		t.Errorf("This is expected IN go test mode!")
	}
}
