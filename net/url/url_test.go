package url_test

import (
	"net/url"
	"testing"

	urlutil "github.com/Andrew-M-C/go.util/net/url"
	"github.com/smartystreets/goconvey/convey"
)

var (
	cv = convey.Convey
	so = convey.So

	eq = convey.ShouldEqual

	isNil = convey.ShouldBeNil
)

func TestURL(t *testing.T) {
	cv("URL.String", t, func() {
		u, err := url.ParseRequestURI("http://www.hh.com/#/#/path?a=A")
		so(err, isNil)

		uu := urlutil.NewURLByOfficial(u)
		so(u.String(), eq, "http://www.hh.com/%23/%23/path?a=A")
		so(uu.String(), eq, "http://www.hh.com/#/#/path?a=A")

		uuu := *uu
		so(uuu.String(), eq, "http://www.hh.com/#/#/path?a=A")
	})
}
