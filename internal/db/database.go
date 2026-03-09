package db

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"time"
)

type Token struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"not null;index"`
	Value     string `gorm:"uniqueIndex;not null"`
	Active    bool   `gorm:"default:true"`
	CreatedAt time.Time
	UpdatedAt time.Time
	LastUsed  *time.Time
	UsageLogs []UsageLog `gorm:"foreignKey:TokenID"`
}

type UsageLog struct {
	ID        uint   `gorm:"primaryKey"`
	TokenID   uint   `gorm:"not null;index"`
	Endpoint  string `gorm:"not null"`
	CreatedAt time.Time
}

func OpenDB(dbPath string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(&Token{}, &UsageLog{})
	if err != nil {
		return nil, err
	}

	return db, nil
}
