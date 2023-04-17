/*
 * @Author: reber
 * @Mail: reber0ask@qq.com
 * @Date: 2022-06-17 11:33:08
 * @LastEditTime: 2023-04-17 08:46:23
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
	"github.com/xuri/excelize/v2"
)

func GetSiteMsg(url string) {
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

	statusCode, title, nowURL, html := httpReq(cloneCtx, targetURL)
	isHttpsScheme1 := strings.Contains(html, "Instead use the HTTPS scheme to access this URL")
	isHttpsScheme2 := strings.Contains(html, "The plain HTTP request was sent to HTTPS port")
	isHttpsScheme3 := strings.Contains(html, "This combination of host and port requires TLS")
	isHttpsScheme4 := strings.Contains(html, "Client sent an HTTP request to an HTTPS server")
	if statusCode == 400 && !(isHttpsScheme1 && isHttpsScheme2 && isHttpsScheme3 && isHttpsScheme4) {
		global.Log.Info(fmt.Sprintf("%s 需要 https 访问", url))
		targetURL = strings.ReplaceAll(targetURL, "http://", "https://")
		targetURL = strings.ReplaceAll(targetURL, ":443/", "/")
		targetURL = strings.ReplaceAll(targetURL, ":80/", "/")
		statusCode, title, nowURL, _ = httpReq(cloneCtx, targetURL)
	}

	global.Log.Info(fmt.Sprintf("[%s] [%d] [%s] %s", url, statusCode, title, nowURL))

	global.Lock.Lock()
	global.Result = append(global.Result, []interface{}{targetURL, statusCode, title, nowURL})
	global.Lock.Unlock()

	// 重新放入 channel
	global.ChTabCtx <- tabCtx
}

func httpReq(cloneCtx context.Context, targetURL string) (int64, string, string, string) {
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
		chromedp.Sleep(time.Duration(global.Opts.WaitTime) * time.Second),
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

func Save2Excel(SheetName string) {
	file := excelize.NewFile()

	// 设置筛选条件
	file.AutoFilter(SheetName, "A1", "D1", "")

	// 设置首行格式
	titleStyle, _ := file.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "left",
		},
	})
	file.SetCellStyle(SheetName, "A1", "D1", titleStyle)

	//设置列宽
	file.SetColWidth(SheetName, "A", "A", 30)
	file.SetColWidth(SheetName, "B", "B", 10)
	file.SetColWidth(SheetName, "C", "C", 30)
	file.SetColWidth(SheetName, "D", "D", 50)

	// 写入首行
	titleSlice := []interface{}{"原始 URL", "状态码", "标题", "当前 URL"}
	_ = file.SetSheetRow(SheetName, "A1", &titleSlice)

	// 写入结果
	for rowID := 0; rowID < len(global.Result); rowID++ {
		rowContent := global.Result[rowID]
		cellName, _ := excelize.CoordinatesToCellName(1, rowID+2) // 从第二行开始写
		if err := file.SetSheetRow(SheetName, cellName, &rowContent); err != nil {
			global.Log.Error(err.Error())
		}
	}

	// 保存工作簿
	if err := file.SaveAs(global.Opts.OutPut); err != nil {
		global.Log.Error(err.Error())
	}

	file.Close()
}
