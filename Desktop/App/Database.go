package main

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type ServerInfo struct {
	ID        uint   `gorm:"primaryKey"`
	IPAddress string `gorm:"not null"`
}

func initDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("serverinfo.db"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Automatically create the "server_infos" table based on the ServerInfo struct
	err = db.AutoMigrate(&ServerInfo{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func fetchIPAddressFromDB(db *gorm.DB) (string, error) {
	var serverInfo ServerInfo

	// Fetch the first record (with ID = 1)
	err := db.First(&serverInfo, 1).Error
	if err != nil {
		return "", err
	}

	return serverInfo.IPAddress, nil
}
