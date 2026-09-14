package repository

import (
	"dana-clone/internal/model"

	"gorm.io/gorm"
)

type MoneyRequestRepository interface {
	Create(req *model.MoneyRequest) error
	FindByID(id uint) (*model.MoneyRequest, error)
	// FindIncoming: permintaan yang DITUJUKAN ke user ini (dia yang harus bayar).
	FindIncoming(userID uint) ([]model.MoneyRequest, error)
	// FindOutgoing: permintaan yang DIBUAT user ini (dia yang menagih orang lain).
	FindOutgoing(userID uint) ([]model.MoneyRequest, error)
	UpdateStatus(id uint, status string) error
}

type moneyRequestRepository struct {
	db *gorm.DB
}

func NewMoneyRequestRepository(db *gorm.DB) MoneyRequestRepository {
	return &moneyRequestRepository{db: db}
}

func (r *moneyRequestRepository) Create(req *model.MoneyRequest) error {
	return r.db.Create(req).Error
}

func (r *moneyRequestRepository) FindByID(id uint) (*model.MoneyRequest, error) {
	var req model.MoneyRequest
	if err := r.db.Preload("Requester").Preload("Target").First(&req, id).Error; err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *moneyRequestRepository) FindIncoming(userID uint) ([]model.MoneyRequest, error) {
	var reqs []model.MoneyRequest
	err := r.db.Preload("Requester").Preload("Target").
		Where("target_id = ?", userID).
		Order("created_at DESC").
		Find(&reqs).Error
	return reqs, err
}

func (r *moneyRequestRepository) FindOutgoing(userID uint) ([]model.MoneyRequest, error) {
	var reqs []model.MoneyRequest
	err := r.db.Preload("Requester").Preload("Target").
		Where("requester_id = ?", userID).
		Order("created_at DESC").
		Find(&reqs).Error
	return reqs, err
}

func (r *moneyRequestRepository) UpdateStatus(id uint, status string) error {
	return r.db.Model(&model.MoneyRequest{}).Where("id = ?", id).Update("status", status).Error
}