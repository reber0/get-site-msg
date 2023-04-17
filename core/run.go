/*
 * @Author: reber
 * @Mail: reber0ask@qq.com
 * @Date: 2022-06-17 15:25:43
 * @LastEditTime: 2023-04-17 08:48:36
 */
package core

import (
	"context"

	"github.com/chromedp/chromedp"
	"github.com/reber0/get-site-msg/global"
)

func Run() {
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
		if targetURL == "" {
			continue
		}

		if global.ChromedpStatus {
			global.Limiter.Take()
			global.WaitGroup.Add()
			go GetSiteMsg(targetURL)
		} else {
			break
		}
	}
	global.WaitGroup.Wait()

	close(global.ChTabCtx)
	for ctx := range global.ChTabCtx {
		ctx.Cancel()
	}

	if global.Opts.OutPut != "" {
		Save2Excel("Sheet1")
	}
}
