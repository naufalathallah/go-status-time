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

    // Create styles
    headerStyle, _ := excel.NewStyle(&excelize.Style{
        Font: &excelize.Font{Bold: true},
        Alignment: &excelize.Alignment{
            Horizontal: "center",
            Vertical: "center",
        },
        Fill: excelize.Fill{
            Type:    "pattern",
            Color:   []string{"#E0E0E0"},
            Pattern: 1,
        },
    })
    excel.SetCellStyle(sheetName, "A1", "B1", headerStyle)

    // Create styles for different row types
    normalStyle, _ := excel.NewStyle(&excelize.Style{
        Alignment: &excelize.Alignment{
            WrapText: true,
            Vertical: "top",
        },
    })

    onLeaveStyle, _ := excel.NewStyle(&excelize.Style{
        Alignment: &excelize.Alignment{
            WrapText: true,
            Vertical: "top",
        },
        Fill: excelize.Fill{
            Type:    "pattern",
            Color:   []string{"#FF0000"},
            Pattern: 1,
        },
    })

    sickStyle, _ := excel.NewStyle(&excelize.Style{
        Alignment: &excelize.Alignment{
            WrapText: true,
            Vertical: "top",
        },
        Fill: excelize.Fill{
            Type:    "pattern",
            Color:   []string{"#FFFF00"},
            Pattern: 1,
        },
    })

    emptyStyle, _ := excel.NewStyle(&excelize.Style{
        Fill: excelize.Fill{
            Type:    "pattern",
            Color:   []string{"#FF0000"},
            Pattern: 1,
        },
    })

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
            
            // Choose style based on content
            styleToApply := normalStyle
            
            if strings.Contains(details, "[ON LEAVE]") {
                styleToApply = onLeaveStyle
            } else if strings.Contains(details, "[SAKIT]") {
                styleToApply = sickStyle
            }
            
            // Apply style to both columns
            excel.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), styleToApply)
            excel.SetCellStyle(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), styleToApply)
        } else {
            // Empty row
            excel.SetCellValue(sheetName, fmt.Sprintf("B%d", row), "")
            excel.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), emptyStyle)
            excel.SetCellStyle(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), emptyStyle)
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