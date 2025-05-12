package crawler_test

import (
	"context"
	"testing"

	"github.com/Andrew-M-C/go.util/crawler"
)

func TestSearchGoogle(t *testing.T) {
	cv("GoogleSearch", t, func() {
		ctx := crawler.NewBrowser(context.Background())
		defer crawler.CloseBrowser(ctx)

		const keywords = "Pakistan J-10CE shoot down Rafale"
		res, err := crawler.GoogleSearch(ctx, keywords,
			crawler.WithDebugger(t.Logf),
			crawler.WithLanguage("阿拉伯语"),
			crawler.WithNum(30),
		)
		so(err, isNil)
		so(len(res), gt, 20)

		for i, hl := range res {
			t.Logf("第 %d 个结果: [%s](%s)", i+1, hl.Title, hl.URL)
		}
	})
}

func TestSearchBing(t *testing.T) {
	cv("Bing Search", t, func() {
		ctx := crawler.NewBrowser(context.Background())
		defer crawler.CloseBrowser(ctx)

		const keywords = "Pakistan J-10CE shoot down Rafale"
		res, err := crawler.BingSearch(ctx, keywords,
			crawler.WithDebugger(t.Logf),
			crawler.WithLanguage("英语"),
			crawler.WithNum(30),
		)
		so(err, isNil)
		so(len(res), gt, 20)

		for i, hl := range res {
			t.Logf("第 %d 个结果: [%s](%s)", i+1, hl.Title, hl.URL)
		}
	})
}
