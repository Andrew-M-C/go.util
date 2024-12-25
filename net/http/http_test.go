package http_test

import (
	"context"
	"encoding/xml"
	"os"
	"path"
	"testing"
	"time"

	"github.com/Andrew-M-C/go-bytesize"
	jsonvalue "github.com/Andrew-M-C/go.jsonvalue"
	"github.com/Andrew-M-C/go.util/net/http"
	"github.com/smartystreets/goconvey/convey"
)

var (
	cv = convey.Convey
	so = convey.So
	eq = convey.ShouldEqual
	gt = convey.ShouldBeGreaterThan

	isNil = convey.ShouldBeNil
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

type testSt struct {
	XMLName xml.Name `xml:"xml" json:"-"` // 指定最外层的类型
	Int     int      `json:"int"`
	Str     string   `json:"str"`
}

func TestJSON(t *testing.T) {
	// reference: https://beeceptor.com/resources/http-echo/
	cv("online echo service for JSON", t, func() {
		req := testSt{
			Int: 9999,
			Str: "xxxx",
		}
		rsp, err := http.JSON[jsonvalue.V](
			context.Background(), "https://echo.free.beeceptor.com",
			http.WithMethod("POST"), http.WithRequestBody(req), http.WithDebugger(t.Logf),
		)
		so(err, isNil)
		t.Log(rsp.MustMarshalString(jsonvalue.OptSetSequence()))

		body, err := rsp.Get("parsedBody")
		so(err, isNil)
		so(jsonvalue.New(req).Equal(body), eq, true)
	})
}

func TestXML(t *testing.T) {
	cv("online echo service for XML", t, func() {
		req := testSt{
			Int: 8888,
			Str: "yyyy",
		}
		b, err := http.XMLGetRspBody(
			context.Background(), "https://echo.free.beeceptor.com",
			http.WithMethod("POST"), http.WithRequestBody(req), http.WithDebugger(t.Logf),
		)
		so(err, isNil)

		t.Logf("%s", b) // 暂时没找到合适的请求和响应都是 XML 的接口
	})
}

func TestDownload(t *testing.T) {
	cv("DownloadFile", t, func() {
		cv("文件名不在路径中", func() {
			const target = "https://cdn.cloudflare.steamstatic.com/client/installer/steam.deb"
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			defer cancel()

			fileName, content, err := http.DownloadFile(
				ctx, target, http.WithDebugger(t.Logf),
				http.WithProgressCallback(func(rp *http.RequestProgress) {
					t.Logf("读取文件 %v / %v", bytesize.Base10(rp.ReadLength()), bytesize.Base10(rp.ContentLength()))
				}),
			)
			so(err, isNil)

			t.Log("下载文件名", fileName)
			t.Log("文件大小", bytesize.Base10(len(content)))
			so(len(content), gt, 0)
			so(path.Ext(fileName), eq, ".deb")
		})
	})
}
