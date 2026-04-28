package model

import (
	"time"
)

type Token struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint64    `gorm:"not null;index;comment:用户ID" json:"user_id"`
	Token     string    `gorm:"type:varchar(512);not null;uniqueIndex;comment:Token值" json:"token"`
	ExpiresAt time.Time `gorm:"not null;comment:过期时间" json:"expires_at"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (Token) TableName() string {
	return "tokens"
}
