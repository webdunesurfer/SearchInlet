package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/webdunesurfer/SearchInlet/internal/db"
	"gorm.io/gorm"
)

type TokenManager struct {
	db          *gorm.DB
	limitPerDay int
}

func NewTokenManager(db *gorm.DB, limitPerDay int) *TokenManager {
	return &TokenManager{
		db:          db,
		limitPerDay: limitPerDay,
	}
}

func GenerateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "sk-" + hex.EncodeToString(bytes), nil
}

func (tm *TokenManager) CreateToken(name string) (*db.Token, error) {
	value, err := GenerateToken()
	if err != nil {
		return nil, err
	}

	token := &db.Token{
		Name:   name,
		Value:  value,
		Active: true,
	}

	if err := tm.db.Create(token).Error; err != nil {
		return nil, err
	}

	return token, nil
}

func (tm *TokenManager) ValidateToken(value string) (*db.Token, error) {
	var token db.Token
	if err := tm.db.Where("value = ? AND active = ?", value, true).First(&token).Error; err != nil {
		return nil, fmt.Errorf("invalid or expired token")
	}

	now := time.Now()
	token.LastUsed = &now
	tm.db.Save(&token)

	return &token, nil
}

func (tm *TokenManager) GetTokenByID(id uint) (*db.Token, error) {
	var token db.Token
	if err := tm.db.First(&token, id).Error; err != nil {
		return nil, err
	}
	return &token, nil
}

func (tm *TokenManager) GetAllTokens() ([]db.Token, error) {
	var tokens []db.Token
	if err := tm.db.Find(&tokens).Error; err != nil {
		return nil, err
	}
	return tokens, nil
}

func (tm *TokenManager) RevokeToken(id uint) error {
	var token db.Token
	if err := tm.db.First(&token, id).Error; err != nil {
		return err
	}
	return tm.db.Model(&token).Update("active", false).Error
}

func (tm *TokenManager) CheckRateLimit(tokenID uint) bool {
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	var count int64
	tm.db.Model(&db.UsageLog{}).Where("token_id = ? AND created_at >= ?", tokenID, todayStart).Count(&count)

	return count < int64(tm.limitPerDay)
}

func (tm *TokenManager) LogUsage(tokenID uint, endpoint string) error {
	log := &db.UsageLog{
		TokenID:  tokenID,
		Endpoint: endpoint,
	}
	return tm.db.Create(log).Error
}

