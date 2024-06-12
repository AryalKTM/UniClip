package clipboard

import (
	database "github.com/AryalKTM/UniClip/Server/Database"
	"github.com/gofiber/fiber/v2"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func GetAllContent(c *fiber.Ctx) error {
	return c.SendString("All Content")
}

func SaveContent(c *fiber.Ctx) error {
	newContent := new(database.clipboardDataStruct)

	err := c.BodyParser(newContent)
	if err != nil {
		c.Status(400).JSON(&fiber.Map{
			"success": false,
			"message": err,
			"data":    nil,
		})
		return err
	}

	result, err := database.CreateEntry(newContent.DataType, newContent.PayloadData, newContent.PostDevice)
	if err != nil {
		c.Status(400).JSON(&fiber.Map{
			"success": false,
			"message": err,
			"data":    nil,
		})
		return err
	}

	c.Status(200).JSON(&fiber.Map{
		"success": false,
		"message": err,
		"data":    nil,
	})
	return nil
}

func CreateEntry(DataType string, PayloadData string, PostDevice string) {
	var newEntry = clipboardDataStruct{
		dataType:    DataType,
		payloadData: PayloadData,
		postDevice:  PostDevice,
	}

	db, err := gorm.Open(sqlite.Open("Clipboard.db"), &gorm.Config{})
	if err != nil {
		return newEntry, err
	}
	db.Create(&clipboardDataStruct{
		DataType: DataType,
	})
}
