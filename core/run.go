/*
 * @Author: reber
 * @Mail: reber0ask@qq.com
 * @Date: 2022-06-17 15:25:43
 * @LastEditTime: 2022-06-18 00:45:42
 */
package core

import (
	"context"
	"gsm/global"

	"github.com/chromedp/chromedp"
)

func Run() {
	// 配置 chrome 选项
	options := []chromedp.ExecAllocatorOption{
		chromedp.Flag("no-default-browser-check", true),   // 启动 chrome 的时候不检查默认浏览器
		chromedp.Flag("headless", global.Opts.IsHeadless), // 是否无头
		chromedp.Flag("no-sandbox", true),                 // 是否关闭沙盒
		chromedp.Flag("mute-audio", false),                // 是否静音
		chromedp.Flag("hide-scrollbars", false),           // 是否隐藏滚动条
		chromedp.UserAgent(`Mozilla/5.0 (Windows NT 10.0; WOW64; rv:78.0) Gecko/20100101 Firefox/78.0`),
		chromedp.WindowSize(1280, 800),
	}
	options = append(chromedp.DefaultExecAllocatorOptions[:], options...)

	// 创建 chrome 窗口
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), options...)
	defer cancel()
	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// 打开一个空 tab
	chromedp.Run(ctx, chromedp.Tasks{
		chromedp.Navigate("about:blank"),
	})

	// 开始获取信息
	for _, url := range global.Targets {
		global.Limiter.Take()
		global.WaitGroup.Add()

		go GetSiteMsg(url, ctx)
	}
	global.WaitGroup.Wait()

	if global.Opts.OutPut != "" {
		Save2Excel("Sheet1")
	}
}
