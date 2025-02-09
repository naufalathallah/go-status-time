package utils

import (
	"fmt"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

func ExportTimesheet(timesheetData string, startDate, endDate time.Time) (*excelize.File, error) {
	excel := excelize.NewFile()
	sheetName := "Timesheet"
	excel.SetSheetName("Sheet1", sheetName)

	// Tambahkan header
	excel.SetCellValue(sheetName, "A1", "Tanggal")
	excel.SetCellValue(sheetName, "B1", "Activity Detail")

	// Pisahkan data berdasarkan baris
	lines := strings.Split(timesheetData, "\n")

	// Simpan data aktivitas per tanggal dalam map
	dataMap := make(map[string]string)
	var lastDate string
	for _, line := range lines {
		// Jika baris adalah tanggal (diakhiri dengan ":")
		if strings.HasSuffix(line, ":") {
			lastDate = strings.TrimSuffix(line, ":")
			dataMap[lastDate] = "" // Inisialisasi tanggal dengan aktivitas kosong
		} else if strings.TrimSpace(line) != "" && lastDate != "" {
			// Baris activity detail (tanpa koma, dipisahkan dengan newline)
			activities := strings.Split(line, ", ")
			details := strings.Join(activities, "\n") // Ubah koma menjadi newline
			dataMap[lastDate] = details
		}
	}

	// Iterasi melalui tanggal dari rentang waktu dan tambahkan ke Excel
	row := 2
	currentDate := startDate
	for !currentDate.After(endDate) {
		dateKey := currentDate.Format("2 January 2006")
		excel.SetCellValue(sheetName, fmt.Sprintf("A%d", row), dateKey)
		if details, exists := dataMap[dateKey]; exists && details != "" {
			// Tambahkan aktivitas jika ada
			excel.SetCellValue(sheetName, fmt.Sprintf("B%d", row), details)
		} else {
			// Jika tidak ada aktivitas, tambahkan dengan kolom kosong
			excel.SetCellValue(sheetName, fmt.Sprintf("B%d", row), "")
		}
		row++
		currentDate = currentDate.Add(24 * time.Hour)
	}

	// Atur lebar kolom agar data terlihat rapi
	excel.SetColWidth(sheetName, "A", "B", 30)

	return excel, nil
}