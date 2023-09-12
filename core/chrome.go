/*
 * @Author: reber
 * @Mail: reber0ask@qq.com
 * @Date: 2022-06-17 11:33:08
 * @LastEditTime: 2023-09-12 10:58:12
 */
package core

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/reber0/get-site-msg/global"
)

func ChromeRun() {
	// 配置 chrome 选项
	options := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("no-default-browser-check", true),        // 启动 chrome 的时候不检查默认浏览器
		chromedp.Flag("headless", global.Opts.IsHeadless),      // 是否无头
		chromedp.Flag("no-sandbox", true),                      // 是否关闭沙盒
		chromedp.Flag("mute-audio", true),                      // 是否静音
		chromedp.Flag("hide-scrollbars", false),                // 是否隐藏滚动条
		chromedp.Flag("ignore-certificate-errors", true),       // 忽略网站证书错误
		chromedp.Flag("blink-settings", "imagesEnabled=false"), // 禁止加载图片
		chromedp.UserAgent(`Mozilla/5.0 (Windows NT 10.0; WOW64; rv:78.0) Gecko/20100101 Firefox/78.0`),
		chromedp.WindowSize(1280, 800),
	)

	// 创建 chrome 窗口
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), options...)
	defer cancel()
	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// 打开一个空 tab
	chromedp.Run(ctx, chromedp.Tasks{
		chromedp.Navigate("about:blank"),
	})

	for i := 0; i < global.Opts.Rate; i++ {
		cloneCtx, cancel := chromedp.NewContext(ctx)

		// open blank tab
		chromedp.Run(cloneCtx, chromedp.Tasks{
			chromedp.Navigate("about:blank"),
		})

		var ctxTmp global.TabCtx
		ctxTmp.Ctx = cloneCtx
		ctxTmp.Cancel = cancel
		global.ChTabCtx <- ctxTmp
	}

	// 开始获取信息
	for _, targetURL := range global.Targets {
		if global.ChromedpStatus {
			global.Limiter.Take()
			global.WaitGroup.Add()
			go WorkerChrome(targetURL)
		} else {
			break
		}
	}
	global.WaitGroup.Wait()

	close(global.ChTabCtx)
	for ctx := range global.ChTabCtx {
		ctx.Cancel()
	}
}

func WorkerChrome(url string) {
	defer global.WaitGroup.Done()

	tabCtx := <-global.ChTabCtx

	// 设置 timeout，用于请求网页的超时
	cloneCtx, cancel := context.WithTimeout(tabCtx.Ctx, time.Duration(global.Opts.TimeOut)*time.Second)
	defer cancel()

	targetURL := strings.TrimRight(url, "/")
	if !strings.HasPrefix(targetURL, "http") {
		targetURL = fmt.Sprintf("http://%s/", targetURL)
	} else {
		targetURL = fmt.Sprintf("%s/", targetURL)
	}
	targetURL = strings.ReplaceAll(targetURL, ":80/", "/")

	statusCode, title, nowURL, html := ChromeReq(cloneCtx, targetURL)
	isHttpsScheme1 := strings.Contains(html, "Instead use the HTTPS scheme to access this URL")
	isHttpsScheme2 := strings.Contains(html, "The plain HTTP request was sent to HTTPS port")
	isHttpsScheme3 := strings.Contains(html, "This combination of host and port requires TLS")
	isHttpsScheme4 := strings.Contains(html, "Client sent an HTTP request to an HTTPS server")
	if statusCode == 400 && !(isHttpsScheme1 && isHttpsScheme2 && isHttpsScheme3 && isHttpsScheme4) {
		// global.Log.Info(fmt.Sprintf("%s 需要 https 访问", url))
		targetURL = strings.ReplaceAll(targetURL, "http://", "https://")
		targetURL = strings.ReplaceAll(targetURL, ":443/", "/")
		targetURL = strings.ReplaceAll(targetURL, ":80/", "/")
		statusCode, title, nowURL, _ = ChromeReq(cloneCtx, targetURL)
	}

	global.Log.Info(fmt.Sprintf("[%s] [%d] [%s] %s", url, statusCode, title, nowURL))

	global.Lock.Lock()
	global.Result = append(global.Result, []interface{}{url, statusCode, title, nowURL})
	global.Lock.Unlock()

	// 重新放入 channel
	global.ChTabCtx <- tabCtx
}

func ChromeReq(cloneCtx context.Context, targetURL string) (int64, string, string, string) {
	var statusCode int64
	var title string
	var nowURL string
	var html string

	// 监听事件，用于获取 StatusCode
	chromedp.ListenTarget(cloneCtx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *network.EventRequestWillBeSent:
			response := ev.RedirectResponse
			if response != nil {
				statusCode = response.Status
			}
		case *network.EventResponseReceived:
			response := ev.Response
			if response.URL == targetURL {
				statusCode = response.Status
			}
		case *page.EventJavascriptDialogOpening:
			fmt.Println("closing dialog:", ev.Message)
			go func() {
				// 自动关闭 Dialog 对话框
				if err := chromedp.Run(cloneCtx,
					// 注释掉下一行可以更清楚地看到效果
					page.HandleJavaScriptDialog(true),
				); err != nil {
					global.Log.Error(err.Error())
				}
			}()
		default:
			// fmt.Println(ev)
		}
	})

	// 请求页面，获取 title
	err := chromedp.Run(cloneCtx, chromedp.Tasks{
		chromedp.Navigate(targetURL),
		chromedp.WaitReady("body", chromedp.ByQuery),
		// chromedp.SendKeys(`input[name=code]`, "3333"),
		// // chromedp.SetValue(`#input_code`, `3333`, chromedp.ByID),
		// chromedp.Click(`/html/body/form/input[2]`, chromedp.BySearch),
		// //在这里加上你需要的后续操作，如Navigate，SendKeys，Click等
		// chromedp.OuterHTML("title", &Title, chromedp.BySearch),
		// 等待页面渲染
		chromedp.Sleep(time.Duration(2) * time.Second),
		chromedp.Location(&nowURL),
		chromedp.Title(&title),
		chromedp.OuterHTML("html", &html),
	})
	if err != nil {
		global.Log.Error(targetURL + " " + err.Error())

		if err.Error() == "context canceled" {
			global.ChromedpStatus = false
		}
	}

	if nowURL == "" {
		nowURL = targetURL
	}

	return statusCode, title, nowURL, html
}
