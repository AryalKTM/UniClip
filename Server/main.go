package main

import (
	"log"
	"github.com/gofiber/fiber/v2"
	"github.com/AryalKTM/UniClip/Server/Clipboard"
	"github.com/AryalKTM/UniClip/Server/Database"
)

func main() {
	// Initialize the database
	if err := database.InitDatabase(); err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello World")
	})

	app.Post("/clipboard", clipboard.SaveContent)
	app.Get("/clipboard", clipboard.GetAllContent)
	app.Get("/clipboard/latest", clipboard.GetLatestContent)
	app.Get("/clipboard/:id", clipboard.GetContentByID)
	app.Put("/clipboard/:id", clipboard.UpdateContent)
	app.Delete("/clipboard/:id", clipboard.DeleteContent)

	log.Fatal(app.Listen(":3000"))
}
