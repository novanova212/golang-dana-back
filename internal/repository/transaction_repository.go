package repository

import (
	"dana-clone/internal/model"

	"gorm.io/gorm"
)

type TransactionRepository interface {
	// Record menerima 'tx' (bisa db biasa, atau tx dari dalam sebuah
	// database transaction) supaya pencatatan riwayat ini BISA ikut
	// di-rollback juga kalau proses Transfer di tengah jalan gagal.
	Record(tx *gorm.DB, transaction *model.Transaction) error
	FindByUserID(userID uint) ([]model.Transaction, error)
}

type transactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) TransactionRepository {
	return &transactionRepository{db: db}
}

func (r *transactionRepository) Record(tx *gorm.DB, transaction *model.Transaction) error {
	// Kalau 'tx' tidak diberikan (nil), pakai koneksi db biasa.
	// Ini berguna untuk kasus TopUp yang tidak butuh transaction.
	if tx == nil {
		tx = r.db
	}
	return tx.Create(transaction).Error
}

// FindByUserID mengambil semua riwayat transaksi milik satu user,
// diurutkan dari yang PALING BARU ke paling lama.
func (r *transactionRepository) FindByUserID(userID uint) ([]model.Transaction, error) {
	var transactions []model.Transaction
	err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&transactions).Error
	if err != nil {
		return nil, err
	}
	return transactions, nil
}