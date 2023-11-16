package runtime_test

import (
	"encoding/json"
	"testing"

	"github.com/Andrew-M-C/go.util/runtime"
	"github.com/smartystreets/goconvey/convey"
)

var (
	cv = convey.Convey
	// so = convey.So
)

func TestRuntime(t *testing.T) {
	cv("GetAllStatic", t, func() { testGetAllStatic(t) })
}

func testGetAllStatic(t *testing.T) {
	a := runtime.GetAllStatic()
	b, _ := json.Marshal(a)
	t.Log(string(b))
}
