package clipboard

import (
	"github.com/AryalKTM/UniClip/Server/Database"
	"github.com/gofiber/fiber/v2"
<<<<<<< HEAD
	"path/filepath"
=======
	"os"
	"strconv"
>>>>>>> 3cd06a7e373914e88712bdc58841324cd9e5064e
)

func GetAllContent(c *fiber.Ctx) error {
	var contents []database.ClipboardData

<<<<<<< HEAD
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
=======
	if err := database.DB.Find(&contents).Error; err != nil {
		return c.Status(500).JSON(&fiber.Map{
>>>>>>> 3cd06a7e373914e88712bdc58841324cd9e5064e
			"success": false,
			"message": err.Error(),
		})
	}

	return c.Status(200).JSON(&fiber.Map{
		"success": true,
		"data":    contents,
	})
}

func GetContentByID(c *fiber.Ctx) error {
	id := c.Params("id")
	contentID, err := strconv.Atoi(id)
	if err != nil {
		return c.Status(400).JSON(&fiber.Map{
			"success": false,
			"message": "Invalid ID",
		})
	}

	var content database.ClipboardData

	if err := database.DB.First(&content, contentID).Error; err != nil {
		return c.Status(404).JSON(&fiber.Map{
			"success": false,
			"message": "Content not found",
		})
	}

	return c.Status(200).JSON(&fiber.Map{
		"success": true,
		"data":    content,
	})
}

func GetLatestContent(c *fiber.Ctx) error {
	var content database.ClipboardData

	// Order by ID in descending order to get the latest item
	if err := database.DB.Order("id desc").First(&content).Error; err != nil {
		return c.Status(404).JSON(&fiber.Map{
			"success": false,
			"message": "Content not found",
		})
	}

	return c.Status(200).JSON(&fiber.Map{
		"success": true,
		"data":    content,
	})
}

func SaveContent(c *fiber.Ctx) error {
	dataType := c.FormValue("dataType")
	payloadData := c.FormValue("payloadData")
	postDevice := c.FormValue("postDevice")
	file, err := c.FormFile("file")

	var fileName string
	var filePath string

	if file != nil {
		fileName = file.Filename
		filePath = "./uploads/" + fileName

		err = c.SaveFile(file, filePath)
		if err != nil {
			return c.Status(500).JSON(&fiber.Map{
				"success": false,
				"message": err.Error(),
			})
		}
	}

	newContent, err := database.CreateEntry(dataType, payloadData, postDevice, fileName, filePath)
	if err != nil {
		return c.Status(500).JSON(&fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	return c.Status(201).JSON(&fiber.Map{
		"success": true,
		"data":    newContent,
	})
}

func UpdateContent(c *fiber.Ctx) error {
	id := c.Params("id")
	contentID, err := strconv.Atoi(id)
	if err != nil {
		return c.Status(400).JSON(&fiber.Map{
			"success": false,
			"message": "Invalid ID",
		})
	}

	var content database.ClipboardData
	if err := database.DB.First(&content, contentID).Error; err != nil {
		return c.Status(404).JSON(&fiber.Map{
			"success": false,
			"message": "Content not found",
		})
	}

	// Update fields
	content.DataType = c.FormValue("dataType")
	content.PayloadData = c.FormValue("payloadData")
	content.PostDevice = c.FormValue("postDevice")

	// Check if a new file is uploaded
	file, err := c.FormFile("file")
	if err == nil {
		// Remove the old file if it exists
		if content.FilePath != "" {
			if err := os.Remove(content.FilePath); err != nil {
				return c.Status(500).JSON(&fiber.Map{
					"success": false,
					"message": "Failed to delete old file",
				})
			}
		}

		// Save the new file
		fileName := file.Filename
		filePath := "./uploads/" + fileName
		if err := c.SaveFile(file, filePath); err != nil {
			return c.Status(500).JSON(&fiber.Map{
				"success": false,
				"message": "Failed to save file",
			})
		}

		// Update file path in content
		content.FileName = fileName
		content.FilePath = filePath
	}

	// Save updated content to database
	if err := database.DB.Save(&content).Error; err != nil {
		return c.Status(500).JSON(&fiber.Map{
			"success": false,
			"message": "Failed to update content",
		})
	}

	return c.Status(200).JSON(&fiber.Map{
		"success": true,
		"data":    content,
	})
}

func DeleteContent(c *fiber.Ctx) error {
	id := c.Params("id")
	contentID, err := strconv.Atoi(id)
	if err != nil {
		return c.Status(400).JSON(&fiber.Map{
			"success": false,
			"message": "Invalid ID",
		})
	}

	var content database.ClipboardData
	if err := database.DB.First(&content, contentID).Error; err != nil {
		return c.Status(404).JSON(&fiber.Map{
			"success": false,
			"message": "Content not found",
		})
	}

	if content.FilePath != "" {
		if err := os.Remove(content.FilePath); err != nil {
			return c.Status(500).JSON(&fiber.Map{
				"success": false,
				"message": err.Error(),
			})
		}
	}

	if err := database.DB.Delete(&content).Error; err != nil {
		return c.Status(500).JSON(&fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	return c.Status(200).JSON(&fiber.Map{
		"success": true,
		"message": "Content deleted successfully",
	})
}
