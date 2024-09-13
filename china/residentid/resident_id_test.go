package residentid_test

import (
	"os"
	"testing"

	"github.com/Andrew-M-C/go.util/china/residentid"
	"github.com/smartystreets/goconvey/convey"
)

var (
	cv = convey.Convey
	so = convey.So
	// eq = convey.ShouldEqual

	isNil = convey.ShouldBeNil
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func TestGeneral(t *testing.T) {
	cv("基础逻辑", t, func() {
		// Reference: https://zhidao.baidu.com/question/25861280.html
		id, err := residentid.New("440102198001021230")
		so(err, isNil)
		t.Log(id.DetailInfo())

		id, err = residentid.New("445121201803163925")
		so(err, isNil)
		t.Log(id.DetailInfo())
	})
}
