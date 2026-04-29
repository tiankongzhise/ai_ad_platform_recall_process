package repository

import (
	"ai_ad_platform_recall_process/internal/model"
	"ai_ad_platform_recall_process/pkg/database"

	"gorm.io/gorm"
)

type RecallRepository struct {
	db *gorm.DB
}

func NewRecallRepository() *RecallRepository {
	return &RecallRepository{
		db: database.GetDB(),
	}
}

func (r *RecallRepository) Create(record *model.RecallRecord) error {
	return r.db.Create(record).Error
}

type QueryParams struct {
	UserName          string
	UID               string
	Platform          string
	UserTag           string
	Page              int
	PageSize          int
}

type QueryResult struct {
	Total   int64              `json:"total"`
	Page    int                `json:"page"`
	PageSize int               `json:"page_size"`
	Records []model.RecallRecord `json:"records"`
}

func (r *RecallRepository) Query(params QueryParams) (*QueryResult, error) {
	query := r.db.Model(&model.RecallRecord{})

	if params.UserName != "" {
		query = query.Where("user_name = ?", params.UserName)
	}
	if params.UID != "" {
		query = query.Where("uid = ?", params.UID)
	}
	if params.Platform != "" {
		query = query.Where("platform = ?", params.Platform)
	}
	if params.UserTag != "" {
		query = query.Where("user_tag = ?", params.UserTag)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 || params.PageSize > 100 {
		params.PageSize = 20
	}

	offset := (params.Page - 1) * params.PageSize

	var records []model.RecallRecord
	if err := query.Order("created_at DESC").Offset(offset).Limit(params.PageSize).Find(&records).Error; err != nil {
		return nil, err
	}

	return &QueryResult{
		Total:   total,
		Page:    params.Page,
		PageSize: params.PageSize,
		Records: records,
	}, nil
}

func (r *RecallRepository) QueryLatest(params QueryParams) (*model.RecallRecord, error) {
	query := r.db.Model(&model.RecallRecord{})

	if params.UserName != "" {
		query = query.Where("user_name = ?", params.UserName)
	}
	if params.UID != "" {
		query = query.Where("uid = ?", params.UID)
	}
	if params.Platform != "" {
		query = query.Where("platform = ?", params.Platform)
	}
	if params.UserTag != "" {
		query = query.Where("user_tag = ?", params.UserTag)
	}

	var record model.RecallRecord
	err := query.Order("created_at DESC").First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *RecallRepository) QueryAll(params QueryParams) ([]model.RecallRecord, error) {
	query := r.db.Model(&model.RecallRecord{})

	if params.UserName != "" {
		query = query.Where("user_name = ?", params.UserName)
	}
	if params.UID != "" {
		query = query.Where("uid = ?", params.UID)
	}
	if params.Platform != "" {
		query = query.Where("platform = ?", params.Platform)
	}
	if params.UserTag != "" {
		query = query.Where("user_tag = ?", params.UserTag)
	}

	var records []model.RecallRecord
	if err := query.Order("created_at DESC").Limit(1000).Find(&records).Error; err != nil {
		return nil, err
	}
	return records, nil
}
