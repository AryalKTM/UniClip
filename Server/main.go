package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/AryalKTM/UniClip/Server/Clipboard"
	"github.com/AryalKTM/UniClip/Server/Database"
	"log"
	"os"
)

func main() {
	// Initialize the database
	if err := database.InitDatabase(); err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	// Create uploads directory if not exists
	if _, err := os.Stat("uploads"); os.IsNotExist(err) {
		err := os.MkdirAll("uploads", os.ModePerm)
		if err != nil {
			log.Fatalf("Failed to create uploads directory: %v", err)
		}
	}

	// Initialize the Fiber app
	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("UniClip is Working")
	})

	app.Post("/clipboard", clipboard.SaveContent)
	app.Get("/clipboard", clipboard.GetAllContent)
	app.Get("/clipboard/latest", clipboard.GetLatestContent)
	app.Get("/clipboard/:id", clipboard.GetContentByID)
	app.Put("/clipboard/:id", clipboard.UpdateContent)
	app.Delete("/clipboard/:id", clipboard.DeleteContent)

	log.Fatal(app.Listen(":3000"))
}
