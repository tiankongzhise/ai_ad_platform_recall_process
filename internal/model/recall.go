package model

import (
	"time"
)

type RecallRecord struct {
	ID                uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	RecallServiceName string    `gorm:"type:varchar(64);not null;index;comment:服务用户名" json:"recall_service_name"`
	Platform          string    `gorm:"type:varchar(64);not null;index;comment:平台来源" json:"platform"`
	UserName          string    `gorm:"type:varchar(128);not null;index;comment:授权用户名称" json:"user_name"`
	Params            string    `gorm:"type:text;comment:完整参数JSON" json:"params,omitempty"`
	CreatedAt         time.Time `gorm:"autoCreateTime;index" json:"created_at"`
}

func (RecallRecord) TableName() string {
	return "recall_records"
}
