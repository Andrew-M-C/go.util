package crawler_test

import (
	"context"
	"testing"
	"time"

	"github.com/Andrew-M-C/go.util/crawler"
)

func TestGetHTMLAndImage(t *testing.T) {
	cv("获取网页和图片", t, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		ctx = crawler.NewBrowser(ctx)
		defer crawler.CloseBrowser(ctx)

		res, err := crawler.GetHTML(
			ctx, "https://www.zhihu.com/question/61678069", crawler.WithDebugger(t.Logf),
		)
		so(err, isNil)
		so(len(res.Images), gt, 0)

		for img := range res.Images {
			t.Log("链接: ", img)
		}
	})
}
