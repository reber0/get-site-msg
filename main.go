/*
 * @Author: reber
 * @Mail: reber0ask@qq.com
 * @Date: 2022-06-16 09:25:13
 * @LastEditTime: 2023-09-12 11:30:10
 */
package main

import (
	"github.com/reber0/get-site-msg/core"
	"github.com/reber0/get-site-msg/entry"
	"github.com/reber0/get-site-msg/global"
)

func main() {
	entry.AppInit()

	if global.Opts.IsChrome {
		core.ChromeRun()
	} else {
		core.RestyRun()
	}

	if global.Opts.OutPut != "" {
		core.Save2Excel("Sheet1")
	}
}
