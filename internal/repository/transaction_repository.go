package repository

import (
	"dana-clone/internal/model"

	"gorm.io/gorm"
)

type TransactionRepository interface {
	Record(tx *gorm.DB, transaction *model.Transaction) error
	// FindByUserID sekarang mendukung pagination. Mengembalikan data
	// halaman yang diminta, PLUS jumlah total baris yang cocok filter
	// (dibutuhkan frontend untuk tahu apakah masih ada data selanjutnya).
	FindByUserID(userID uint, txType string, search string, page int, limit int) ([]model.Transaction, int64, error)
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

func (r *transactionRepository) FindByUserID(userID uint, txType string, search string, page int, limit int) ([]model.Transaction, int64, error) {
	query := r.db.Model(&model.Transaction{}).Where("user_id = ?", userID)

	if txType != "" {
		query = query.Where("type = ?", txType)
	}
	if search != "" {
		// ILIKE khusus PostgreSQL (case-insensitive). Untuk driver lain
		// (misal SQLite yang dipakai saat unit test), pakai LIKE biasa
		// supaya query tetap valid meski hasilnya sedikit lebih ketat
		// soal huruf besar/kecil.
		if r.db.Dialector.Name() == "postgres" {
			query = query.Where("description ILIKE ?", "%"+search+"%")
		} else {
			query = query.Where("description LIKE ?", "%"+search+"%")
		}
	}

	// Hitung total dulu SEBELUM di-limit, supaya frontend tahu total
	// data yang tersedia (dipakai untuk tombol "muat lebih banyak").
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit

	var transactions []model.Transaction
	err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&transactions).Error
	if err != nil {
		return nil, 0, err
	}

	return transactions, total, nil
}