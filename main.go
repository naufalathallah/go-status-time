package main

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/naufalathallah/go-status-time/handlers"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("⚠️ .env file not found or failed to load")
	}
}

func main() {
	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Welcome")
	})

	app.Post("/weekly", handlers.WeeklyHandler)
	app.Post("/timesheet", handlers.TimesheetHandler)
	app.Post("/timesheet-worklog", handlers.TimesheetWorklogHandler)

	fmt.Println("Server berjalan di http://localhost:8000")
	if err := app.Listen(":8000"); err != nil {
		log.Fatalf("Error: %v", err)
	}
}
