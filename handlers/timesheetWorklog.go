package handlers

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func TimesheetWorklogHandler(c *fiber.Ctx) error {

	fmt.Println("=== Timesheet Data ===")
	return c.SendString("Timesheet Worklog Handler is not implemented yet")
}
