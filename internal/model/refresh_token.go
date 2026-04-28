package model

import (
	"time"
)

type RefreshToken struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint64    `gorm:"not null;index;comment:用户ID" json:"user_id"`
	Token     string    `gorm:"type:varchar(128);not null;uniqueIndex;comment:RefreshToken值" json:"token"`
	ExpiresAt time.Time `gorm:"not null;comment:过期时间" json:"expires_at"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (RefreshToken) TableName() string {
	return "refresh_tokens"
}
