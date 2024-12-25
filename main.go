package main

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/naufalathallah/go-status-time/handlers"
)

func main() {
	app := fiber.New()

	app.Post("/upload", handlers.UploadHandler)

	fmt.Println("Server berjalan di http://localhost:8000")
	if err := app.Listen(":8000"); err != nil {
		log.Fatalf("Error: %v", err)
	}
}