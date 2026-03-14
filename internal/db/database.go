package db

import (
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
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

type LoginAttempt struct {
	ID        uint      `gorm:"primaryKey"`
	IP        string    `gorm:"index;not null"`
	Success   bool      `gorm:"not null"`
	CreatedAt time.Time `gorm:"index"`
}

type GlobalSetting struct {
	ID        uint      `gorm:"primaryKey"`
	Key       string    `gorm:"uniqueIndex;not null"`
	Value     string    `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func OpenDB(dbPath string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(&Token{}, &UsageLog{}, &LoginAttempt{}, &GlobalSetting{})
	if err != nil {
		return nil, err
	}

	return db, nil
}
