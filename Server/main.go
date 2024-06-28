package main

import (
	"github.com/AryalKTM/UniClip/Server/Clipboard"
	"github.com/AryalKTM/UniClip/Server/Database"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"log"
)

func main() {
	// Initialize the database
	if err := database.InitDatabase(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize the Fiber app
	app := fiber.New()

	// Middleware
	app.Use(logger.New())

	// Routes
	app.Get("/clipboard", clipboard.GetAllContent)
	app.Post("/clipboard", clipboard.SaveContent)

	// Start server
	log.Fatal(app.Listen(":3000"))
}
