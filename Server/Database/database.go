package database

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type clipboardDataStruct struct {
	gorm.Model
	DataType    string `json:"dataType"`
	PayloadData string `json:"payloadData"`
	PostDevice  string `json:"postDevice"`
}

func InitDatabase() error {
	db, err := gorm.Open(sqlite.Open("Clipboard.db"), &gorm.Config{})

	if err != nil {
		return err
	}

	db.AutoMigrate(&clipboardDataStruct{})

	return nil
}
