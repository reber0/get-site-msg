/*
 * @Author: reber
 * @Mail: reber0ask@qq.com
 * @Date: 2022-06-17 11:30:03
 * @LastEditTime: 2023-04-17 08:48:19
 */
package entry

import (
	"os"

	"github.com/reber0/get-site-msg/global"
	"github.com/reber0/go-common/mylog"
	"github.com/reber0/go-common/utils"
	"github.com/remeh/sizedwaitgroup"
	"go.uber.org/ratelimit"
)

// 初始化全局变量、解析参数等初始化的操作
func AppInit() {
	// 初始化部分参数
	global.RootPath, _ = os.Getwd()
	global.TermWidth = utils.GetTermWidth()
	global.Log = mylog.New().IsToFile(true).Logger()

	// 解析参数
	ParseOptions()

	global.Limiter = ratelimit.New(global.Opts.Rate)
	global.WaitGroup = sizedwaitgroup.New(global.Opts.Rate)

	global.ChromedpStatus = true
	global.ChTabCtx = make(chan global.TabCtx, global.Opts.Rate)

	global.Targets = utils.FileEachLineRead(global.Opts.TargetFile)
	global.Result = make([][]interface{}, 0, len(global.Targets))
}
