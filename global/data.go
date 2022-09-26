/*
 * @Author: reber
 * @Mail: reber0ask@qq.com
 * @Date: 2022-06-17 11:30:35
 * @LastEditTime: 2022-09-26 13:42:54
 */
package global

import (
	"sync"

	"github.com/remeh/sizedwaitgroup"
	"go.uber.org/ratelimit"
	"go.uber.org/zap"
)

var Opts struct {
	Version string

	TargetFile string
	Rate       int
	TimeOut    int
	WaitTime   int
	OutPut     string
	IsHeadless bool
}

var (
	RootPath  string
	TermWidth int // 终端宽度
	Log       *zap.Logger

	Limiter   ratelimit.Limiter             // 控制执行 Worker 的频率
	WaitGroup sizedwaitgroup.SizedWaitGroup // 控制总的并发数

	Lock sync.Mutex

	ChromedpStatus bool // Chromedp 状态

	Targets []string
	Result  [][]interface{}
)
