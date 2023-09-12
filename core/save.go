/*
 * @Author: reber
 * @Mail: reber0ask@qq.com
 * @Date: 2023-09-12 09:03:08
 * @LastEditTime: 2023-09-12 09:03:30
 */
package core

import (
	"github.com/reber0/get-site-msg/global"
	"github.com/xuri/excelize/v2"
)

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
