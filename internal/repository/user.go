package repository

import (
	"ai_ad_platform_recall_process/internal/model"
	"ai_ad_platform_recall_process/pkg/database"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository() *UserRepository {
	return &UserRepository{
		db: database.GetDB(),
	}
}

func (r *UserRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepository) FindByUsername(username string) (*model.User, error) {
	var user model.User
	err := r.db.Where("recall_service_name = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByID(id uint64) (*model.User, error) {
	var user model.User
	err := r.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) UpdateNotifyURL(userID uint64, notifyURL string) error {
	return r.db.Model(&model.User{}).Where("id = ?", userID).Update("notify_url", notifyURL).Error
}

func (r *UserRepository) FindByPhone(phone string) (*model.User, error) {
	var user model.User
	err := r.db.Where("phone = ?", phone).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) UpdatePassword(userID uint64, hashedPassword string) error {
	return r.db.Model(&model.User{}).Where("id = ?", userID).Update("password", hashedPassword).Error
}

func (r *UserRepository) UpdatePhone(userID uint64, phone string) error {
	return r.db.Model(&model.User{}).Where("id = ?", userID).Update("phone", phone).Error
}

func (r *UserRepository) Delete(userID uint64) error {
	return r.db.Model(&model.User{}).Where("id = ?", userID).Update("status", 0).Error
}

func (r *UserRepository) UpdateApiToken(userID uint64, apiToken string) error {
	return r.db.Model(&model.User{}).Where("id = ?", userID).Update("api_token", apiToken).Error
}

func (r *UserRepository) FindByApiToken(apiToken string) (*model.User, error) {
	var user model.User
	err := r.db.Where("api_token = ?", apiToken).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindActiveByPhone(phone string) (*model.User, error) {
	var user model.User
	err := r.db.Unscoped().Where("phone = ? AND status = 1", phone).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
