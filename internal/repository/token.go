package repository

import (
	"ai_ad_platform_recall_process/internal/model"
	"ai_ad_platform_recall_process/pkg/database"

	"gorm.io/gorm"
)

type TokenRepository struct {
	db *gorm.DB
}

func NewTokenRepository() *TokenRepository {
	return &TokenRepository{
		db: database.GetDB(),
	}
}

func (r *TokenRepository) Create(token *model.Token) error {
	return r.db.Create(token).Error
}

func (r *TokenRepository) FindByToken(tokenStr string) (*model.Token, error) {
	var token model.Token
	err := r.db.Where("token = ?", tokenStr).First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *TokenRepository) DeleteByToken(tokenStr string) error {
	return r.db.Where("token = ?", tokenStr).Delete(&model.Token{}).Error
}

func (r *TokenRepository) DeleteByUserID(userID uint64) error {
	return r.db.Where("user_id = ?", userID).Delete(&model.Token{}).Error
}

func (r *TokenRepository) DeleteExpiredTokens() error {
	return r.db.Where("expires_at < NOW()").Delete(&model.Token{}).Error
}

func (r *TokenRepository) FindByUserID(userID uint64) ([]model.Token, error) {
	var tokens []model.Token
	err := r.db.Where("user_id = ?", userID).Find(&tokens).Error
	if err != nil {
		return nil, err
	}
	return tokens, nil
}
