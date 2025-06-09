package utils

import (
	"fmt"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

// ExportTimesheetWorklog creates an Excel file from timesheet worklog data
func ExportTimesheetWorklog(timesheetData string, startDate, endDate time.Time) (*excelize.File, error) {
    excel := excelize.NewFile()
    sheetName := "Worklog Timesheet"
    excel.SetSheetName("Sheet1", sheetName)

    // Add headers
    excel.SetCellValue(sheetName, "A1", "Tanggal")
    excel.SetCellValue(sheetName, "B1", "Activity Detail")

    // Split data by lines
    lines := strings.Split(timesheetData, "\n")

    // Store activities by date in a map
    dataMap := make(map[string]string)
    var lastDate string
    for _, line := range lines {
        // If line is a date (ends with ":")
        if strings.HasSuffix(line, ":") {
            lastDate = strings.TrimSuffix(line, ":")
            dataMap[lastDate] = "" // Initialize date with empty activity
        } else if strings.TrimSpace(line) != "" && lastDate != "" {
            // Line is activity detail
            activities := strings.Split(line, ", ")
            details := strings.Join(activities, "\n") // Convert commas to newlines
            dataMap[lastDate] = details
        }
    }

    // Fill in Excel with data from date range
    row := 2
    currentDate := startDate
    for !currentDate.After(endDate) {
        dateKey := currentDate.Format("2 January 2006")
        excel.SetCellValue(sheetName, fmt.Sprintf("A%d", row), dateKey)
        
        if details, exists := dataMap[dateKey]; exists && details != "" {
            // Add the formatted activities to the cell
            excel.SetCellValue(sheetName, fmt.Sprintf("B%d", row), details)
            
            // Set style for wrapping text
            style, _ := excel.NewStyle(&excelize.Style{
                Alignment: &excelize.Alignment{
                    WrapText: true,
                    Vertical: "top",
                },
            })
            excel.SetCellStyle(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), style)
        } else {
            excel.SetCellValue(sheetName, fmt.Sprintf("B%d", row), "")
        }
        row++
        currentDate = currentDate.Add(24 * time.Hour)
    }

    // Set column width for better readability
    excel.SetColWidth(sheetName, "A", "A", 20)
    excel.SetColWidth(sheetName, "B", "B", 80)

    // Set row height to accommodate multiple lines
    for i := 2; i <= row-1; i++ {
        excel.SetRowHeight(sheetName, i, 80)
    }

    return excel, nil
}