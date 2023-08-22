package base64_test

import (
	"bytes"
	"encoding/base64"
	"strings"
	"testing"

	base64util "github.com/Andrew-M-C/go.util/encoding/base64"
	"github.com/smartystreets/goconvey/convey"
)

var (
	cv = convey.Convey
	so = convey.So
	eq = convey.ShouldEqual

	isNil  = convey.ShouldBeNil
	notNil = convey.ShouldNotBeNil
)

func TestBase64(t *testing.T) {
	cv("基本逻辑", t, func() { testBase64(t) })
}

func testBase64(t *testing.T) {
	cv("不需要补齐 =", func() {
		raw := []byte{1, 2, 3, 4, 5, 6}
		enc := base64.StdEncoding.EncodeToString(raw)
		t.Logf("Got base64 encoded: %s", enc)
		so(strings.HasSuffix(enc, "="), eq, false)

		dec, err := base64util.StdEncoding.DecodeString(enc)
		so(err, isNil)
		so(bytes.Compare(dec, raw), eq, 0)
	})

	cv("需要补齐 =", func() {
		raw := []byte{1, 2, 3, 4}
		enc := base64.StdEncoding.EncodeToString(raw)
		t.Logf("Got base64 encoded: %s", enc)
		so(strings.HasSuffix(enc, "="), eq, true)

		utilEnc := base64util.StdEncoding.EncodeToString(raw)
		so(utilEnc, eq, enc)

		enc = strings.TrimRight(enc, "=")
		so(strings.HasSuffix(enc, "="), eq, false)

		_, err := base64.StdEncoding.DecodeString(enc)
		so(err, notNil)

		dec, err := base64.StdEncoding.DecodeString(enc + "==")
		so(err, isNil)
		so(bytes.Compare(dec, raw), eq, 0)

		utilDec, err := base64util.StdEncoding.DecodeString(enc)
		so(err, isNil)
		so(bytes.Compare(utilDec, raw), eq, 0)
	})
}
