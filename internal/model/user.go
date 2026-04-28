package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID                   uint64     `gorm:"primaryKey;autoIncrement" json:"-"`
	UserName             string     `gorm:"column:user_name;type:varchar(64);not null;uniqueIndex;comment:用户名" json:"user_name"`
	UID                  string     `gorm:"column:uid;type:varchar(64);uniqueIndex;comment:用户唯一标识UID" json:"uid"`
	Phone                string     `gorm:"type:varchar(20);comment:手机号" json:"phone,omitempty"`
	Password             string     `gorm:"type:varchar(255);not null;comment:加密密码" json:"-"`
	ApiToken             string     `gorm:"type:varchar(64);uniqueIndex;comment:用户API Token(长期有效)" json:"api_token,omitempty"`
	NotifyURL            string     `gorm:"type:varchar(512);default:'';comment:通知回调URL" json:"notify_url,omitempty"`
	Status               int8       `gorm:"type:tinyint;default:1;comment:状态 1正常 0已注销" json:"status"`
	CreatedAt            time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt            time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt            gorm.DeletedAt `gorm:"index" json:"-"`
}

func (User) TableName() string {
	return "users"
}
