package http_test

import (
	"context"
	"os"
	"testing"

	jsonvalue "github.com/Andrew-M-C/go.jsonvalue"
	"github.com/Andrew-M-C/go.util/net/http"
	"github.com/smartystreets/goconvey/convey"
)

var (
	cv = convey.Convey
	so = convey.So
	eq = convey.ShouldEqual

	isNil = convey.ShouldBeNil
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

type testSt struct {
	Int int    `json:"int"`
	Str string `json:"str"`
}

func TestJSON(t *testing.T) {
	// reference: https://beeceptor.com/resources/http-echo/
	cv("online echo service", t, func() {
		req := testSt{
			Int: 9999,
			Str: "xxxx",
		}
		var rsp *jsonvalue.V
		rsp, err := http.JSON[*jsonvalue.V](
			context.Background(), "https://echo.free.beeceptor.com",
			http.WithMethod("POST"), http.WithRequestBody(req),
		)
		so(err, isNil)
		t.Log(rsp.MustMarshalString(jsonvalue.OptSetSequence()))

		body, err := rsp.Get("parsedBody")
		so(err, isNil)
		so(jsonvalue.New(req).Equal(body), eq, true)
	})
}
