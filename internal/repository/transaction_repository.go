package repository

import (
	"dana-clone/internal/model"

	"gorm.io/gorm"
)

type TransactionRepository interface {
	Record(tx *gorm.DB, transaction *model.Transaction) error
	// FindByUserID: txType kosong "" berarti semua tipe, search kosong ""
	// berarti tidak filter berdasarkan kata kunci deskripsi.
	FindByUserID(userID uint, txType string, search string) ([]model.Transaction, error)
}

type transactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) TransactionRepository {
	return &transactionRepository{db: db}
}

func (r *transactionRepository) Record(tx *gorm.DB, transaction *model.Transaction) error {
	if tx == nil {
		tx = r.db
	}
	return tx.Create(transaction).Error
}

func (r *transactionRepository) FindByUserID(userID uint, txType string, search string) ([]model.Transaction, error) {
	query := r.db.Where("user_id = ?", userID)

	if txType != "" {
		query = query.Where("type = ?", txType)
	}
	if search != "" {
		// ILIKE = pencarian case-insensitive khusus PostgreSQL
		// (mirip LIKE tapi tidak peduli huruf besar/kecil).
		query = query.Where("description ILIKE ?", "%"+search+"%")
	}

	var transactions []model.Transaction
	err := query.Order("created_at DESC").Find(&transactions).Error
	if err != nil {
		return nil, err
	}
	return transactions, nil
}