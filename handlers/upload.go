package handlers

import (
	"fmt"
	"sort"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/naufalathallah/go-status-time/utils"
)

func UploadHandler(c *fiber.Ctx) error {
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

	groupedData := utils.GroupByColumnData(headers, records, startDate, endDate)

	keys := make([]string, 0, len(groupedData))
	for key := range groupedData {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	response := fmt.Sprintf("startDate: %s\nendDate: %s\n\n", startDateStr, endDateStr)
	response += "Hasil Pengelompokan Data:\n"

	for _, key := range keys {
		data := groupedData[key]
		response += fmt.Sprintf("Key: %s\n", key)

		for _, col := range []string{"Issue Type", "Summary", "Status", "Assignee", "Labels", "Story point estimate", "Created", "Bloked-GQA"} {
			if val, ok := data[col]; ok {
				response += fmt.Sprintf("  %s: %v\n", col, val)
			}
		}

		if details, ok := data["Details"].(map[string][]string); ok {
			for column, values := range details {
				response += fmt.Sprintf("  %s:\n", column)
				for _, value := range values {
					response += fmt.Sprintf("    %s\n", value)
				}
			}
		}
		response += "\n"
	}
	fmt.Println(response)

	excelFile, err := utils.CreateExcelFile(groupedData)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Gagal membuat file Excel")
	}

	buffer, err := excelFile.WriteToBuffer()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Gagal menyimpan file Excel")
	}

	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", "attachment; filename=data.xlsx")
	return c.Send(buffer.Bytes())
}
