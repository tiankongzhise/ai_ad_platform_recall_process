package repository

import (
	"ai_ad_platform_recall_process/internal/model"
	"ai_ad_platform_recall_process/pkg/database"

	"gorm.io/gorm"
)

type RefreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository() *RefreshTokenRepository {
	return &RefreshTokenRepository{
		db: database.GetDB(),
	}
}

func (r *RefreshTokenRepository) Create(token *model.RefreshToken) error {
	return r.db.Create(token).Error
}

func (r *RefreshTokenRepository) FindByToken(tokenStr string) (*model.RefreshToken, error) {
	var token model.RefreshToken
	err := r.db.Where("token = ?", tokenStr).First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *RefreshTokenRepository) DeleteByToken(tokenStr string) error {
	return r.db.Where("token = ?", tokenStr).Delete(&model.RefreshToken{}).Error
}

func (r *RefreshTokenRepository) DeleteByUserID(userID uint64) error {
	return r.db.Where("user_id = ?", userID).Delete(&model.RefreshToken{}).Error
}

func (r *RefreshTokenRepository) DeleteExpiredTokens() error {
	return r.db.Where("expires_at < NOW()").Delete(&model.RefreshToken{}).Error
}

func (r *RefreshTokenRepository) FindByUserID(userID uint64) ([]model.RefreshToken, error) {
	var tokens []model.RefreshToken
	err := r.db.Where("user_id = ?", userID).Find(&tokens).Error
	if err != nil {
		return nil, err
	}
	return tokens, nil
}
