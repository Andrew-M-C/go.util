package deepcopy

import (
	"testing"

	jsonvalue "github.com/Andrew-M-C/go.jsonvalue"
	"github.com/Andrew-M-C/go.util/deepcopy/testcase"
	"github.com/smartystreets/goconvey/convey"
)

var (
	cv = convey.Convey
	so = convey.So

	isNil = convey.ShouldBeNil
)

func TestDeepCopy(t *testing.T) {
	cv("普通调试", t, func() { testDebug(t) })
}

func testDebug(t *testing.T) {
	code, err := BuildDeepCopy(
		// http.Cookie{},
		// &http.Client{},
		(*jsonvalue.V)(nil),
		// time.Time{},
		// &testcase.PointerSlice{},
		// &testcase.ID{},
		&testcase.MapInStruct{},
	).EnableDebug().WithLogFunc(t.Logf).Do()

	so(err, isNil)
	t.Log("codes:", code)

	// TODO:
}
