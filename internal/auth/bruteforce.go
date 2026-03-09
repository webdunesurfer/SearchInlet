package auth

import (
	"time"

	"github.com/webdunesurfer/SearchInlet/internal/db"
	"gorm.io/gorm"
)

const (
	MaxFailedAttempts = 5
	BanDuration       = 1 * time.Hour
	AttemptWindow     = 15 * time.Minute
)

type LoginLimiter struct {
	db *gorm.DB
}

func NewLoginLimiter(db *gorm.DB) *LoginLimiter {
	return &LoginLimiter{db: db}
}

// IsBanned checks if an IP is currently banned from logging in
func (l *LoginLimiter) IsBanned(ip string) bool {
	var count int64
	// Check for failures in the last 15 minutes
	windowStart := time.Now().Add(-AttemptWindow)
	l.db.Model(&db.LoginAttempt{}).
		Where("ip = ? AND success = ? AND created_at > ?", ip, false, windowStart).
		Count(&count)

	return count >= MaxFailedAttempts
}

// LogAttempt records a login attempt
func (l *LoginLimiter) LogAttempt(ip string, success bool) {
	attempt := &db.LoginAttempt{
		IP:      ip,
		Success: success,
	}
	l.db.Create(attempt)

	// Periodic cleanup of old successful attempts or very old failed ones
	if time.Now().Unix()%100 == 0 {
		cutoff := time.Now().Add(-24 * time.Hour)
		l.db.Where("created_at < ?", cutoff).Delete(&db.LoginAttempt{})
	}
}
