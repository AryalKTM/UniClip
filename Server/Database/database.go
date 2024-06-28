package database

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type ClipboardData struct {
	gorm.Model
	DataType    string `json:"dataType"`
	PayloadData string `json:"payloadData"`
	PostDevice  string `json:"postDevice"`
	FileName    string `json:"fileName"`
	FilePath    string `json:"filePath"`
}

var DB *gorm.DB

func InitDatabase() error {
	var err error
	DB, err = gorm.Open(sqlite.Open("clipboard.db"), &gorm.Config{})
	if err != nil {
		return err
	}

	DB.AutoMigrate(&ClipboardData{})
	return nil
}

func CreateEntry(dataType, payloadData, postDevice, fileName, filePath string) (*ClipboardData, error) {
	newEntry := ClipboardData{
		DataType:    dataType,
		PayloadData: payloadData,
		PostDevice:  postDevice,
		FileName:    fileName,
		FilePath:    filePath,
	}

	if err := DB.Create(&newEntry).Error; err != nil {
		return nil, err
	}

	return &newEntry, nil
}
