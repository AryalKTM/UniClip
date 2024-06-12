package main

import (
	"github.com/AryalKTM/UniClip/Server/Clipboard"
	"github.com/AryalKTM/UniClip/Server/Database"
	"github.com/gofiber/fiber/v2"
)

func status(c *fiber.Ctx) error {
	return c.SendString("Server is Running! Send your Request")
}

func setupRoutes(app *fiber.App) {
	app.Get("/", status)
	app.Get("/api/content", clipboard.GetAllContent)
	app.Post("/api/content", clipboard.SaveContent)
}

func main() {
	app := fiber.New()
	
	dbErr := database.InitDatabase()
	if dbErr != nil {
		panic(dbErr)
	}

	setupRoutes(app)
	app.Listen(":3000")
}
