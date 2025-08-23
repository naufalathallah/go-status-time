package handlers

import (
	"github.com/gofiber/fiber/v2"
)

func TimesheetHandler(c *fiber.Ctx) error {

	return c.SendString("v3")
}