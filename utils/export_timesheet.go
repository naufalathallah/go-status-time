package utils

import (
	"fmt"
	"strings"

	"github.com/xuri/excelize/v2"
)

func ExportTimesheet(timesheetData string) (*excelize.File, error) {
	excel := excelize.NewFile()
	sheetName := "Timesheet"
	excel.SetSheetName("Sheet1", sheetName)

	// Tambahkan header
	excel.SetCellValue(sheetName, "A1", "Tanggal")
	excel.SetCellValue(sheetName, "B1", "Activity Detail")

	// Pisahkan data berdasarkan baris
	lines := strings.Split(timesheetData, "\n")
	row := 2

	for _, line := range lines {
		// Jika baris adalah tanggal (diakhiri dengan ":")
		if strings.HasSuffix(line, ":") {
			date := strings.TrimSuffix(line, ":")
			excel.SetCellValue(sheetName, fmt.Sprintf("A%d", row), date)
		} else if strings.TrimSpace(line) != "" {
			// Baris activity detail (tanpa koma, dipisahkan dengan newline)
			activities := strings.Split(line, ", ")
			details := strings.Join(activities, "\n") // Ubah koma menjadi newline
			excel.SetCellValue(sheetName, fmt.Sprintf("B%d", row), details)
			row++
		}
	}

	// Atur lebar kolom agar data terlihat rapi
	excel.SetColWidth(sheetName, "A", "B", 30)

	return excel, nil
}