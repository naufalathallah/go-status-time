package handlers

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/naufalathallah/go-status-time/utils"
)

func UploadHandler(c *fiber.Ctx) error {
	if c.Method() != fiber.MethodPost {
		return c.Status(fiber.StatusMethodNotAllowed).SendString("Hanya menerima metode POST")
	}

	startDate := c.FormValue("startDate")
	endDate := c.FormValue("endDate")

	if startDate == "" || endDate == "" {
		return c.Status(fiber.StatusBadRequest).SendString("startDate dan endDate harus diisi")
	}

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

	groupedData := utils.GroupByColumnData(headers, records)

	response := fmt.Sprintf("startDate: %s\nendDate: %s\n\n", startDate, endDate)
	response += "Hasil Pengelompokan Data:\n"

	for key, columns := range groupedData {
		response += fmt.Sprintf("Key: %s\n", key)
		for column, values := range columns {
			response += fmt.Sprintf("  %s:\n", column)
			for _, value := range values {
				response += fmt.Sprintf("    %s\n", value)
			}
		}
		response += "\n"
	}

	return c.SendString(response)
}

