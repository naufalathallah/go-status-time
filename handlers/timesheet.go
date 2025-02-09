package handlers

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/naufalathallah/go-status-time/utils"
)

func TimesheetHandler(c *fiber.Ctx) error {
	if c.Method() != fiber.MethodPost {
		return c.Status(fiber.StatusMethodNotAllowed).SendString("Hanya menerima metode POST")
	}

	startDateStr := c.FormValue("startDate")
	endDateStr := c.FormValue("endDate")

	if startDateStr == "" || endDateStr == "" {
		return c.Status(fiber.StatusBadRequest).SendString("startDate dan endDate harus diisi")
	}

	startDate, err := time.Parse("2006/01/02", startDateStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Format startDate tidak valid")
	}

	endDate, err := time.Parse("2006/01/02", endDateStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Format endDate tidak valid")
	}
	endDate = endDate.Add(24*time.Hour - time.Second)

	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Gagal mendapatkan file")
	}

	f, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Gagal membuka file")
	}
	defer f.Close()

	headers, records, err := utils.ParseCSV(f)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Gagal membaca file CSV")
	}

	// Logic baru untuk pemrosesan timesheet
	timesheetData := processTimesheetData(headers, records, startDate, endDate)

	fmt.Println("=== Timesheet Data ===")
	fmt.Println(timesheetData)

	excelFile, err := utils.ExportTimesheet(timesheetData, startDate, endDate)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Gagal membuat file Excel")
	}

	// Simpan file Excel ke buffer
	buffer, err := excelFile.WriteToBuffer()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Gagal menyimpan file Excel")
	}

	// Set header dan kirim file Excel
	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=timesheet-%s-to-%s.xlsx", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")))
	return c.Send(buffer.Bytes())
}

func processTimesheetData(headers []string, records [][]string, startDate, endDate time.Time) string {
	// Pemetaan kolom aktivitas ke nama header
	columnPairs := map[string]string{
		"Code Fixing-GQA":         "'->Code Fixing-GQA",
		"Code Review-GQA":         "'->Code Review-GQA",
		"Done-GQA":                "'->Done-GQA",
		"In Progress-GQA":         "'->In Progress-GQA",
		"on hold-GQA":             "'->on hold-GQA",
	}

	// Indeks dari kolom yang relevan
	indexMap := make(map[string]int)
	for key, pair := range columnPairs {
		for i, header := range headers {
			if header == key || header == pair {
				indexMap[header] = i
			}
		}
	}

	// Indeks untuk kolom tambahan (Summary)
	summaryIndex := -1
	for i, header := range headers {
		if header == "Summary" {
			summaryIndex = i
		}
	}

	// Hasil akhir
	result := make(map[string][]string)

	// Proses setiap record
	for _, record := range records {
		if summaryIndex == -1 || summaryIndex >= len(record) {
			continue
		}

		// Ambil judul tugas dari kolom "Summary"
		summary := record[summaryIndex]

		for activity, pairColumn := range columnPairs {
			mainIdx, mainOk := indexMap[activity]
			pairIdx, pairOk := indexMap[pairColumn]

			if mainOk && pairOk && mainIdx < len(record) && pairIdx < len(record) {
				mainValues := strings.Split(record[mainIdx], ",")
				pairValues := strings.Split(record[pairIdx], ",")

				for i := range mainValues {
					if i < len(pairValues) {
						date, err := time.Parse("2006-01-02 15:04", strings.TrimSpace(pairValues[i]))
						if err != nil {
							continue
						}

						// Tambahkan mapping "Summary - Aktivitas"
						if (date.After(startDate) || date.Equal(startDate)) && (date.Before(endDate) || date.Equal(endDate)) {
							dateKey := date.Format("2 January 2006") // Format tanggal untuk output
							mappedValue := fmt.Sprintf("%s - %s", summary, strings.Replace(activity, "-GQA", "", -1))
							result[dateKey] = append(result[dateKey], mappedValue)
						}
					}
				}
			}
		}
	}

	// Buat output teks berdasarkan tanggal
	var output strings.Builder
	currentDate := startDate
	for currentDate.Before(endDate) || currentDate.Equal(endDate) {
		dateKey := currentDate.Format("2 January 2006")
		output.WriteString(dateKey + ":\n")

		if activities, ok := result[dateKey]; ok {
			// Gunakan map untuk menghindari duplikasi
			uniqueActivities := make(map[string]bool)
			var filteredActivities []string

			for _, activity := range activities {
				if !uniqueActivities[activity] {
					uniqueActivities[activity] = true
					filteredActivities = append(filteredActivities, activity)
				}
			}

			// Tambahkan aktivitas yang unik ke output
			if len(filteredActivities) > 0 {
				output.WriteString(strings.Join(filteredActivities, ", ") + "\n")
			}
		}
		currentDate = currentDate.Add(24 * time.Hour)
	}

	return output.String()
}