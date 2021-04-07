package gotest

import (
	"os"
	"strings"
)

// IsInGoTest tells whether we are now in go test mode.
//
// Reference: [How do I know I'm running within “go test”](https://stackoverflow.com/questions/14249217)
func IsInGoTest() bool {
	return strings.HasSuffix(os.Args[0], ".test")
}
