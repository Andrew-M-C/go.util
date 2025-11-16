package crawler

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync/atomic"
	"time"

	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"golang.org/x/net/html"
)

// NewBrowser 新建一个浏览器
func NewBrowser(ctx context.Context) context.Context {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		// 打开无头模式
		chromedp.Flag("headless", true),
		// 设置用户代理，模拟浏览器访问
		chromedp.Flag("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) "+
			"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36"),
		// 设置浏览器窗口大小
		chromedp.WindowSize(1150, 1000),
		// 设置语言
		chromedp.Flag("lang", "zh-CN"),
		// 防止监测webdriver
		chromedp.Flag("enable-automation", false),
		// 禁用blink特征，减少自动化检测
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		// 忽略证书错误
		chromedp.Flag("ignore-certificate-errors", true),
		// 关闭浏览器声音
		chromedp.Flag("mute-audio", false),
		// 再次设置浏览器窗口大小，确保覆盖默认值
		chromedp.WindowSize(1150, 1000),
		// 禁用图片加载
		chromedp.Flag("blink-settings", "imagesEnabled=false"),
		// 禁用视频预加载
		chromedp.Flag("autoplay-policy", "user-gesture-required"),
	)

	ctx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	cancler := &canceler{
		fn:     cancel,
		closed: &atomic.Bool{},
	}
	return context.WithValue(ctx, cancelerKey{}, cancler)
}

// CloseBrowser 关闭浏览器
func CloseBrowser(ctx context.Context) {
	v := ctx.Value(cancelerKey{})
	if v == nil {
		return
	}
	c, ok := v.(*canceler)
	if !ok || c == nil {
		return
	}
	if c.closed.CompareAndSwap(false, true) {
		c.fn()
	}
}

type cancelerKey struct{}

type canceler struct {
	fn     context.CancelFunc
	closed *atomic.Bool
}

func isBrowser(ctx context.Context) bool {
	v := ctx.Value(cancelerKey{})
	if v == nil {
		return false
	}
	c, ok := v.(*canceler)
	if !ok {
		return false
	}
	return c != nil && !c.closed.Load()
}

// HTMLResult 存储HTML内容和Cookie信息
type HTMLResult struct {
	Content string              // HTML内容
	Cookies []*network.Cookie   // Cookie信息
	Images  map[string]struct{} // 引用的所有图片链接
}

// GetHTML 下载 html 静态内容和网站设置的cookie
func GetHTML(ctx context.Context, targetURL string, opts ...Option) (*HTMLResult, error) {
	o := mergeOptions(opts...)

	// 确保这是经过 NewBrowser 创建的浏览器
	if !isBrowser(ctx) {
		return nil, errors.New("请先使用 NewBrowser 创建浏览器")
	}

	// 存储渲染后的 HTML
	var htmlContent string
	var cookies []*network.Cookie
	images := map[string]struct{}{} // 存储所有图片链接

	// 用于执行具体的浏览器操作，设置日志级别为 ERROR 以过滤掉警告
	ctx, cancel := chromedp.NewContext(ctx, chromedp.WithLogf(func(format string, args ...interface{}) {
		if strings.Contains(format, "unhandled page event") {
			return // 忽略未处理的页面事件警告
		}
		o.debug(format, args...)
	}))
	defer cancel() // 确保在函数结束时释放资源

	// 执行浏览器操作
	start := time.Now()
	err := chromedp.Run(ctx,
		// 添加事件监听器
		chromedp.ActionFunc(func(ctx context.Context) error {
			chromedp.ListenTarget(ctx, func(ev interface{}) {
				switch e := ev.(type) {
				case *page.EventFrameStartedNavigating: //nolint:typecheck
					o.debug("开始导航: 到 %s", e.URL)
				case *page.EventFrameNavigated:
					o.debug("导航完成: %s (加载方式: %s)", e.Frame.URL, e.Frame.LoaderID)
				case *page.EventFrameStoppedLoading:
					o.debug("导航停止: %s", e.FrameID)
				case *page.EventLoadEventFired:
					o.debug("页面加载完成，耗时: %v", time.Since(start))
				case *page.EventDomContentEventFired:
					o.debug("DOM内容加载完成, 耗时: %v", time.Since(start))
				default:
					// 记录其他类型的导航事件
					if ev != nil {
						o.debug("其他导航事件: %T", ev)
					}
				}
			})
			return nil
		}),
		chromedp.Navigate(targetURL),
		// 同时处理超时和正常流程
		chromedp.ActionFunc(func(ctx context.Context) error {
			select {
			case <-ctx.Done():
				// 超时触发：停止加载并捕获当前 HTML
				_ = chromedp.Evaluate("window.stop()", nil).Do(ctx)
				// 使用 DOM 方法获取 HTML（兼容原有逻辑）
				node, _ := dom.GetDocument().Do(ctx)
				if node != nil {
					html, _ := dom.GetOuterHTML().WithNodeID(node.NodeID).Do(ctx)
					htmlContent = html
				}
				// 获取Cookie
				cookies, _ = network.GetCookies().Do(ctx)
			default:
				// 正常流程：等待 5 秒后获取 HTML
				_ = chromedp.Sleep(5 * time.Second).Do(ctx)
				node, err := dom.GetDocument().Do(ctx)
				if err != nil {
					return err
				}
				html, err := dom.GetOuterHTML().WithNodeID(node.NodeID).Do(ctx)
				if err != nil {
					return fmt.Errorf("GetOuterHTML().WithNodeID(%v) 失败 (%w)", node.NodeID, err)
				}
				htmlContent = html

				// 获取Cookie
				cookies, err = network.GetCookies().Do(ctx)
				if err != nil {
					return fmt.Errorf("获取Cookie失败 (%w)", err)
				}
			}
			return nil
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("执行 chrome 操作失败 (%w)", err)
	}

	if len(targetURL) > 50 {
		o.debug("耗时 %v - %s...", time.Since(start), targetURL[:50])
	} else {
		o.debug("耗时 %v - %s", time.Since(start), targetURL)
	}

	// 提取所有图片链接
	if htmlContent != "" {
		extractedImages, err := extractImageLinks(htmlContent, targetURL)
		if err != nil {
			o.debug("提取图片链接失败: %v", err)
		} else {
			images = extractedImages
			o.debug("提取到 %d 个图片链接", len(images))
		}
	}

	return &HTMLResult{
		Content: htmlContent,
		Cookies: cookies,
		Images:  images,
	}, nil
}

// ExtractText 提取出所有文本
func ExtractText(htmlBody string) (string, error) {
	// 比较粗糙地解析 HTML 并提取链接, 尝试解析出下一页
	rawDoc, err := html.Parse(strings.NewReader(htmlBody))
	if err != nil {
		return "", fmt.Errorf("解析响应失败 (%w)", err)
	}

	buff := &bytes.Buffer{}
	findText(rawDoc, buff)
	return buff.String(), nil
}

func findText(n *html.Node, buff *bytes.Buffer) {
	// 首先是节点本身
	switch n.Type {
	case html.TextNode:
		if s := n.Data; s != "" {
			buff.WriteRune('\n')
			buff.WriteString(n.Data)
		}

	case html.ElementNode: // 如果是节点类型, 那么要判断一下是否应该跳过
		if isExcludedTag(n.Data) {
			return // 全部跳过, 包括子节点
		}

	default:
		// 什么都不做
	}

	// 然后是子节点
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		findText(child, buff)
	}
}

// 判断是否需要排除的标签
func isExcludedTag(tagName string) bool {
	excludedTags := []string{"script", "style", "nav", "footer", "header", "link", "meta", "aside"}
	for _, tag := range excludedTags {
		if tag == tagName {
			return true
		}
	}
	return false
}

// extractImageLinks 提取 HTML 中所有图片的完整链接
func extractImageLinks(htmlContent, baseURL string) (map[string]struct{}, error) {
	images := make(map[string]struct{})

	// 解析 HTML
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return images, fmt.Errorf("解析 HTML 失败: %w", err)
	}

	// 解析基础 URL
	base, err := url.Parse(baseURL)
	if err != nil {
		return images, fmt.Errorf("解析基础 URL 失败: %w", err)
	}

	// 递归遍历 HTML 节点，查找所有 img 标签
	var findImages func(*html.Node)
	findImages = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "img" {
			// 查找 src 属性
			for _, attr := range n.Attr {
				if attr.Key == "src" || attr.Key == "data-src" || attr.Key == "data-original" {
					imgURL := strings.TrimSpace(attr.Val)
					if imgURL == "" {
						continue
					}

					// 跳过 data URI 和特殊协议
					if strings.HasPrefix(imgURL, "data:") ||
						strings.HasPrefix(imgURL, "javascript:") ||
						strings.HasPrefix(imgURL, "about:") {
						continue
					}

					// 将相对链接转换为绝对链接
					absoluteURL := resolveURL(base, imgURL)
					if absoluteURL != "" {
						images[absoluteURL] = struct{}{}
					}
				}
			}
		}

		// 递归处理子节点
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			findImages(child)
		}
	}

	findImages(doc)
	return images, nil
}

// resolveURL 将相对 URL 转换为绝对 URL
func resolveURL(base *url.URL, href string) string {
	href = strings.TrimSpace(href)
	if href == "" {
		return ""
	}

	// 解析 href
	u, err := url.Parse(href)
	if err != nil {
		return ""
	}

	// 如果已经是绝对 URL，直接返回
	if u.IsAbs() {
		return u.String()
	}

	// 否则基于 base URL 解析
	resolved := base.ResolveReference(u)
	return resolved.String()
}
