package utils

import (
	"sort"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

func CreateExcelFile(groupedData map[string]map[string]interface{}) (*excelize.File, error) {
	excel := excelize.NewFile()
	sheetName := "Data"
	excel.SetSheetName("Sheet1", sheetName)

	header := []string{
		"Key", "Issue Type", "Summary", "Status", "Assignee", "Labels", "Story point estimate", "Created",
		"Bloked-GQA", "Code Fixing-GQA", "Code Review-GQA", "Done-GQA", "In Progress-GQA", "on hold-GQA",
		"Reject Code Review-GQA", "To Do-GQA",
	}
	for i, h := range header {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		excel.SetCellValue(sheetName, cell, h)
	}

	keys := make([]string, 0, len(groupedData))
	for key := range groupedData {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	row := 2
	for _, key := range keys {
		data := groupedData[key]
		colValues := make(map[string]float64)

		if details, ok := data["Details"].(map[string][]string); ok {
			for col, values := range details {
				for _, value := range values {
					parts := strings.Split(value, " ")
					num, _ := strconv.ParseFloat(strings.ReplaceAll(parts[0], ",", "."), 64)
					colValues[col] += num
				}
			}
		}

		cols := []interface{}{
			key,
			data["Issue Type"], data["Summary"], data["Status"], data["Assignee"], data["Labels"],
			data["Story point estimate"], data["Created"],
			"-", // Placeholder untuk "Bloked-GQA"
			formatNumber(colValues["Code Fixing-GQA"]),
			formatNumber(colValues["Code Review-GQA"]),
			formatNumber(colValues["Done-GQA"]),
			formatNumber(colValues["In Progress-GQA"]),
			formatNumber(colValues["on hold-GQA"]),
			formatNumber(colValues["Reject Code Review-GQA"]),
			formatNumber(colValues["To Do-GQA"]),
		}
		for i, val := range cols {
			cell, _ := excelize.CoordinatesToCellName(i+1, row)
			excel.SetCellValue(sheetName, cell, val)
		}
		row++
	}

	return excel, nil
}
