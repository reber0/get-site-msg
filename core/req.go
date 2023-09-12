/*
 * @Author: reber
 * @Mail: reber0ask@qq.com
 * @Date: 2023-09-12 09:04:09
 * @LastEditTime: 2023-09-12 10:58:19
 */
package core

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/reber0/get-site-msg/global"
)

func RestyRun() {
	// 开始获取信息
	for _, targetURL := range global.Targets {
		global.Limiter.Take()
		global.WaitGroup.Add()
		go WorkerResty(targetURL)
	}
	global.WaitGroup.Wait()
}

func WorkerResty(url string) {
	defer global.WaitGroup.Done()

	targetURL := strings.TrimRight(url, "/")
	if !strings.HasPrefix(targetURL, "http") {
		targetURL = fmt.Sprintf("http://%s/", targetURL)
	} else {
		targetURL = fmt.Sprintf("%s/", targetURL)
	}
	targetURL = strings.ReplaceAll(targetURL, ":80/", "/")

	statusCode, title, nowURL, html := RestyReq(targetURL)
	isHttpsScheme1 := strings.Contains(html, "Instead use the HTTPS scheme to access this URL")
	isHttpsScheme2 := strings.Contains(html, "The plain HTTP request was sent to HTTPS port")
	isHttpsScheme3 := strings.Contains(html, "This combination of host and port requires TLS")
	isHttpsScheme4 := strings.Contains(html, "Client sent an HTTP request to an HTTPS server")
	if statusCode == 400 && !(isHttpsScheme1 && isHttpsScheme2 && isHttpsScheme3 && isHttpsScheme4) {
		// global.Log.Info(fmt.Sprintf("%s 需要 https 访问", url))
		targetURL = strings.ReplaceAll(targetURL, "http://", "https://")
		targetURL = strings.ReplaceAll(targetURL, ":443/", "/")
		targetURL = strings.ReplaceAll(targetURL, ":80/", "/")
		statusCode, title, nowURL, _ = RestyReq(targetURL)
	}

	global.Log.Info(fmt.Sprintf("%s ==> [%d][%s] %s", url, statusCode, title, nowURL))

	global.Lock.Lock()
	global.Result = append(global.Result, []interface{}{url, statusCode, title, nowURL})
	global.Lock.Unlock()
}

func RestyReq(targetURL string) (int, string, string, string) {
	resp, err := global.Client.R().Get(targetURL)
	if err != nil {
		// global.Log.Error(err.Error())
		return 0, "", "", ""
	}
	statusCode := resp.StatusCode()
	nowURL := resp.RawResponse.Request.URL.String()
	html := string(resp.Body())

	dom, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		global.Log.Error(err.Error())
	}
	title := dom.Find("title").Text()

	return statusCode, title, nowURL, html
}
