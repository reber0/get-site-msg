/*
 * @Author: reber
 * @Mail: reber0ask@qq.com
 * @Date: 2022-06-16 09:25:13
 * @LastEditTime: 2023-04-17 08:49:05
 */
package main

import (
	"github.com/reber0/get-site-msg/core"
	"github.com/reber0/get-site-msg/entry"
)

func main() {
	entry.AppInit()

	core.Run()
}
