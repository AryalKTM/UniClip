package clipboard

import (
	"github.com/AryalKTM/UniClip/Server/Database"
	"github.com/gofiber/fiber/v2"
	"path/filepath"
)

func GetAllContent(c *fiber.Ctx) error {
	var contents []database.ClipboardData
	database.DB.Find(&contents)
	return c.JSON(contents)
}

func SaveContent(c *fiber.Ctx) error {
	newContent := new(database.ClipboardData)

	if form, err := c.MultipartForm(); err == nil {
		files := form.File["file"]
		for _, file := range files {
			// Save the file to the server
			filePath := filepath.Join("Server/uploads", file.Filename)
			if err := c.SaveFile(file, filePath); err != nil {
				return c.Status(500).SendString(err.Error())
			}

			newContent.FileName = file.Filename
			newContent.FilePath = filePath
		}
	}

	if err := c.BodyParser(newContent); err != nil {
		return c.Status(400).JSON(&fiber.Map{
			"success": false,
			"message": err.Error(),
			"data":    nil,
		})
	}

	result, err := database.CreateEntry(newContent.DataType, newContent.PayloadData, newContent.PostDevice, newContent.FileName, newContent.FilePath)
	if err != nil {
		return c.Status(400).JSON(&fiber.Map{
			"success": false,
			"message": err.Error(),
			"data":    nil,
		})
	}

	return c.Status(200).JSON(&fiber.Map{
		"success": true,
		"message": "Content saved successfully",
		"data":    result,
	})
}
