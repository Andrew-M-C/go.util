package runtime_test

import (
	"encoding/json"
	"net"
	"testing"

	"github.com/Andrew-M-C/go.util/runtime"
	"github.com/smartystreets/goconvey/convey"
)

var (
	cv = convey.Convey
	so = convey.So

	isTrue = convey.ShouldBeTrue
)

func TestRuntime(t *testing.T) {
	cv("GetAllStatic", t, func() { testGetAllStatic(t) })
	cv("IPIsLAN", t, func() { testIPIsLAN(t) })
}

func testGetAllStatic(t *testing.T) {
	a := runtime.GetAllStatic()
	b, _ := json.Marshal(a)
	t.Log(string(b))
}

func testIPIsLAN(t *testing.T) {
	ip := net.ParseIP("fe80::b09c:4aff:febf:92a2")
	so(runtime.IPIsLAN(ip), isTrue)
}
