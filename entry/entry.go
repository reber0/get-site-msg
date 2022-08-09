/*
 * @Author: reber
 * @Mail: reber0ask@qq.com
 * @Date: 2022-06-17 11:30:03
 * @LastEditTime: 2022-07-22 11:01:00
 */
package entry

import (
	"gsm/global"
	"os"

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
	global.Log = mylog.New().Logger()

	// 解析参数
	ParseOptions()

	global.Limiter = ratelimit.New(global.Opts.Rate)
	global.WaitGroup = sizedwaitgroup.New(global.Opts.Rate)

	global.Targets = utils.FileEachLineRead(global.Opts.TargetFile)
	global.Result = make([][]interface{}, 0, len(global.Targets))
}
