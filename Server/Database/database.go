package database

import (
	"path/filepath"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"os"
)

var DB *gorm.DB

type ClipboardData struct {
	gorm.Model
	DataType    string `json:"dataType"`
	PayloadData string `json:"payloadData"`
	PostDevice  string `json:"postDevice"`
	FileName    string `json:"fileName"`
	FilePath    string `json:"filePath"`
}

func InitDatabase() error {
	// Ensure the Server/Database directory exists
	err := os.MkdirAll("Server/Database", os.ModePerm)
	if err != nil {
		return err
	}

	// Set the database path
	dbPath := filepath.Join("Server/Database", "Clipboard.db")

	// Open the SQLite database
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return err
	}

	// Auto migrate the ClipboardData struct
	return DB.AutoMigrate(&ClipboardData{})
}

func CreateEntry(dataType, payloadData, postDevice, fileName, filePath string) (ClipboardData, error) {
	newEntry := ClipboardData{
		DataType:    dataType,
		PayloadData: payloadData,
		PostDevice:  postDevice,
		FileName:    fileName,
		FilePath:    filePath,
	}

	if err := DB.Create(&newEntry).Error; err != nil {
		return ClipboardData{}, err
	}

	return newEntry, nil
}
