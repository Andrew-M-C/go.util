package crawler

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// GoogleSearch 执行 Google 搜索
func GoogleSearch(
	ctx context.Context, keywords string, opts ...Option,
) ([]HyperLink, error) {
	o := mergeGoogleSearchOptions(opts...)
	body, err := getGoogleBody(ctx, keywords, o)
	if err != nil {
		return nil, err
	}
	// o.debug("%s", body)
	res, err := parseGoogleBody(body)
	if err != nil {
		return nil, err
	}
	return deDuplicateHyperLinks(res), nil
}

func getGoogleBody(ctx context.Context, keywords string, o *options) (string, error) {
	keywords = strings.TrimSpace(keywords)

	const targetURL = "https://google.com.hk/search"

	q := url.Values{}
	q.Add("q", keywords)
	if o.language != "" {
		q.Add("lr", o.language)
	}
	if o.num > 0 {
		q.Add("num", strconv.Itoa(o.num))
	}

	u, _ := url.Parse(targetURL)
	u.RawQuery = q.Encode()
	finalTargetURL := u.String()
	o.debug("target URL: '%s'", finalTargetURL)

	start := time.Now()
	res, err := GetHTML(ctx, finalTargetURL)
	ela := time.Since(start)
	if err != nil {
		return "", err
	}

	o.debug("搜索 '%s' 耗时 %v", keywords, ela)
	return res.Content, nil
}

func parseGoogleBody(body string) ([]HyperLink, error) {
	doc, err := html.Parse(bytes.NewBufferString(body))
	if err != nil {
		return nil, fmt.Errorf("解析响应失败 (%w)", err)
	}

	results := traverseNodes(nil, doc)
	return results, nil
}

func traverseNodes(nodePath []string, n *html.Node) []HyperLink {
	if n == nil {
		return nil
	}

	var res []HyperLink

	// 仅处理元素节点（忽略文本、注释等）, 同时也仅处理 超链接
	if n.Type == html.ElementNode && n.Data == "a" {
		// if len(nodePath) > 0 {
		// 	log.Debugf("%v 元素标签: <%s>", nodePath, n.Data)
		// } else {
		// 	log.Debugf("元素标签: <%s>", n.Data)
		// }

		// 如果是超链接, 那么将文本打出来
		title, link := "", ""

		// 优先查找aria-label属性
		for _, attr := range n.Attr {
			if attr.Key == "aria-label" && attr.Val != "" {
				title = attr.Val
				break
			}
		}

		// 如果没有找到aria-label，则使用原有逻辑提取文本
		if title == "" && n.Data == "a" {
			title = extractText(n)
			title = strings.TrimSpace(title)
		}

		// 打印所有属性
		for _, attr := range n.Attr {
			if attr.Key == "" && attr.Val == "" {
				continue
			}
			// log.Debugf("  └─ 属性: %s = %s", attr.Key, attr.Val)
			if attr.Key == "href" {
				link = attr.Val
			}
		}

		if title != "" && link != "" && !strings.HasSuffix(title, link) {
			hl := parseHyperLink(title, link)
			if hl.URL != "" {
				res = append(res, hl)
			}
		}
	}

	// 递归遍历子节点
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		if n.Data == "" {
			res = append(res, traverseNodes(nil, child)...)
		} else {
			res = append(res, traverseNodes(append(nodePath, n.Data), child)...)
		}
	}

	return res
}

func parseHyperLink(title, link string) HyperLink {
	if !strings.HasPrefix(link, "http") {
		return HyperLink{}
	}
	u, err := url.Parse(link)
	if err != nil {
		return HyperLink{}
	}
	if isExcludedHostInGoogleSearch(u.Host) {
		return HyperLink{}
	}
	// 不知道为什么搜出来的链接 TEXT 会包含 URL, 因此这里去掉
	if i := strings.Index(title, u.Host); i >= 0 {
		title = title[:i]
	}
	// Google 搜出来的会带 https:// 后缀, 也去掉
	title = strings.TrimSuffix(title, "https://")
	title = strings.TrimSuffix(title, "http://")
	// 返回
	return HyperLink{
		Title: title,
		URL:   link,
	}
}

func isExcludedHostInGoogleSearch(host string) bool {
	excludedHosts := map[string]bool{
		"www.beian.gov.cn":      true,
		"go.microsoft.com":      true,
		"support.microsoft.com": true,
		"beian.miit.gov.cn":     true,
		"dxzhgl.miit.gov.cn":    true,
		"support.google.com":    true,
		"policies.google.com":   true,
		"accounts.google.com":   true,
		"maps.google.com.hk":    true,
		"www.google.com.hk":     true,
		"translate.google.com":  true,
	}
	return excludedHosts[host]
}

// 提取节点下所有文本内容（包括嵌套文本）
func extractText(n *html.Node) string {
	var text strings.Builder

	// 递归遍历子节点
	for curr := n.FirstChild; curr != nil; curr = curr.NextSibling {
		if curr.Type == html.TextNode {
			text.WriteString(curr.Data)
		} else if curr.Type == html.ElementNode {
			text.WriteString(extractText(curr)) // 递归处理嵌套标签
		}
	}

	return text.String()
}

func mergeGoogleSearchOptions(opts ...Option) *options {
	opt := mergeOptions(opts...)

	// 特殊参数翻译
	if opt.language != "" {
		if lang, exist := googleLangs[opt.language]; exist {
			opt.language = "lang_" + lang
		}
	}
	return opt
}

var googleLangs = map[string]string{
	"阿拉伯语":   "ar",
	"保加利亚语":  "bg",
	"加泰罗尼亚语": "ca",
	"克罗地亚语":  "hr",
	"中文(简体)": "zh-Hans",
	"中文简体":   "zh-Hans",
	"简体中文":   "zh-Hans",
	"简中":     "zh-Hans",
	"中文(繁体)": "zh-Hant",
	"中文繁体":   "zh-Hant",
	"繁体中文":   "zh-Hant",
	"繁中":     "zh-Hant",
	"捷克语":    "cs",
	"丹麦语":    "da",
	"荷兰语":    "nl",
	"英语":     "en",
	"菲律宾语":   "fil",
	"芬兰语":    "fi",
	"法语":     "fr",
	"德语":     "de",
	"希腊语":    "el",
	"希伯来语":   "he",
	"印地语":    "hi",
	"匈牙利语":   "hu",
	"印尼语":    "id",
	"意大利语":   "it",
	"日语":     "ja",
	"韩语":     "ko",
	"拉脱维亚语":  "lv",
	"立陶宛语":   "lt",
	"挪威语":    "no",
	"波兰语":    "pl",
	"葡萄牙语":   "pt",
	"罗马尼亚语":  "ro",
	"俄语":     "ru",
	"塞尔维亚语":  "sr",
	"斯洛伐克语":  "sk",
	"斯洛文尼亚语": "sl",
	"西班牙语":   "es",
	"瑞典语":    "sv",
	"泰语":     "th",
	"土耳其语":   "tr",
	"乌克兰语":   "uk",
	"越南语":    "vi",
}

func deDuplicateHyperLinks(links []HyperLink) []HyperLink {
	res := make([]HyperLink, 0, len(links))
	added := make(map[string]struct{}, len(links))
	for _, hl := range links {
		if _, exist := added[hl.URL]; exist {
			continue
		}
		res = append(res, hl)
		added[hl.URL] = struct{}{}
	}
	return res
}
