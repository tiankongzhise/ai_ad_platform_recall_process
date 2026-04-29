package model

import (
	"time"
)

type RecallRecord struct {
	ID                uint64    `gorm:"primaryKey;autoIncrement" json:"-"`
	UserName          string    `gorm:"column:user_name;type:varchar(64);not null;index;comment:用户名" json:"user_name"`
	UID               string    `gorm:"column:uid;type:varchar(64);index;comment:用户UID(唯一身份标识)" json:"uid,omitempty"`
	Platform          string    `gorm:"type:varchar(64);not null;index;comment:平台来源" json:"platform"`
	UserTag           string    `gorm:"column:user_tag;type:varchar(128);not null;index;comment:授权用户标识" json:"user_tag"`
	Params            string    `gorm:"type:text;comment:完整参数JSON" json:"params,omitempty"`
	CreatedAt         time.Time `gorm:"autoCreateTime;index" json:"created_at"`
}

func (RecallRecord) TableName() string {
	return "recall_records"
}
