/*
 * @Author: reber
 * @Mail: reber0ask@qq.com
 * @Date: 2022-06-17 21:28:11
 * @LastEditTime: 2022-06-21 13:04:10
 */
package entry

import (
	"flag"
	"fmt"
	"gsm/global"
	"os"
	"strings"

	"github.com/reber0/go-common/utils"
)

//解析命令行参数，得到所有参数信息
func ParseOptions() {
	global.Opts.Version = "0.01"

	flag.Usage = usage // 改变默认的 Usage
	flag.StringVar(&global.Opts.TargetFile, "iL", "", "指定目标文件")
	flag.IntVar(&global.Opts.Rate, "r", 10, "扫描速率")
	flag.IntVar(&global.Opts.TimeOut, "t", 10, "超时时间")
	flag.IntVar(&global.Opts.WaitTime, "w", 2, "等待时间(等待页面渲染的时间)")
	flag.BoolVar(&global.Opts.IsHeadless, "show", false, "是否使用无头模式 (default false)")
	flag.StringVar(&global.Opts.OutPut, "O", "", "将结果保存到 xlsx 文件")

	flag.Parse() // 通过调用 flag.Parse() 来对命令行参数进行解析

	checkOption()
}

func usage() {
	fmt.Print(`Usage: gsm [-h] [-iL TargetFile] [-O OutPutFileName]
		[-r Rate] [-t TimeOut] [-w WaitTime]

Options:
`)
	flag.PrintDefaults() // 调用 PrintDefaults 打印前面定义的参数列表。
}

// 检查参数
func checkOption() {
	if global.Opts.TargetFile == "" {
		global.Log.Error("Missing required parameter -iL\n")
		flag.Usage()
		os.Exit(0)
	}
	if !utils.IsFileExist(global.Opts.TargetFile) {
		msg := fmt.Sprintf("目标文件 %s 不存在 !", global.Opts.TargetFile)
		global.Log.Error(msg)
		os.Exit(0)
	}
	if global.Opts.Rate > 20 {
		global.Log.Error("Rate 最大不能超过 20 !")
		os.Exit(0)
	}
	if global.Opts.OutPut != "" && !strings.HasSuffix(global.Opts.OutPut, ".xlsx") {
		global.Opts.OutPut = global.Opts.OutPut + ".xlsx"
	}
}
