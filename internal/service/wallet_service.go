package service

import (
	"errors"

	"dana-clone/internal/model"
	"dana-clone/internal/repository"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type WalletService interface {
	TopUp(userID uint, amount int64) (int64, error)
	Transfer(fromUserID, toUserID uint, amount int64) error
	TransferWithNote(fromUserID, toUserID uint, amount int64, senderNote, receiverNote string) error
	GetHistory(userID uint, txType string, search string, page int, limit int) ([]model.Transaction, int64, error)
}

type walletService struct {
	userRepo repository.UserRepository
	txRepo   repository.TransactionRepository
	db       *gorm.DB
}

func NewWalletService(userRepo repository.UserRepository, txRepo repository.TransactionRepository, db *gorm.DB) WalletService {
	return &walletService{userRepo: userRepo, txRepo: txRepo, db: db}
}

func (s *walletService) TopUp(userID uint, amount int64) (int64, error) {
	if amount <= 0 {
		return 0, errors.New("jumlah top up harus lebih dari 0")
	}

	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return 0, errors.New("user tidak ditemukan")
	}

	newBalance := user.Balance + amount
	if err := s.userRepo.UpdateBalance(userID, newBalance); err != nil {
		return 0, err
	}

	s.txRepo.Record(nil, &model.Transaction{
		UserID:      userID,
		Type:        "topup",
		Amount:      amount,
		Description: "Top up saldo",
	})

	return newBalance, nil
}

func (s *walletService) Transfer(fromUserID, toUserID uint, amount int64) error {
	return s.TransferWithNote(fromUserID, toUserID, amount, "Transfer keluar", "Transfer masuk")
}

// TransferWithNote adalah logic INTI perpindahan saldo, aman dari race
// condition lewat database transaction + row locking.
//
// Catatan soal locking: clause.Locking (SELECT ... FOR UPDATE) hanya
// didukung oleh database seperti PostgreSQL/MySQL, TIDAK oleh SQLite.
// Supaya kode ini tetap bisa diuji dengan SQLite in-memory (lebih cepat,
// tanpa perlu database beneran saat unit test) TANPA mengorbankan
// keamanan di production, locking hanya diaktifkan kalau driver-nya
// benar-benar PostgreSQL.
func (s *walletService) TransferWithNote(fromUserID, toUserID uint, amount int64, senderNote, receiverNote string) error {
	if amount <= 0 {
		return errors.New("jumlah transfer harus lebih dari 0")
	}

	if fromUserID == toUserID {
		return errors.New("tidak bisa transfer ke diri sendiri")
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		usesLocking := tx.Dialector.Name() == "postgres"

		senderQuery := tx
		if usesLocking {
			senderQuery = tx.Clauses(clause.Locking{Strength: "UPDATE"})
		}
		var sender model.User
		if err := senderQuery.First(&sender, fromUserID).Error; err != nil {
			return errors.New("pengirim tidak ditemukan")
		}

		if sender.Balance < amount {
			return errors.New("saldo tidak cukup")
		}

		receiverQuery := tx
		if usesLocking {
			receiverQuery = tx.Clauses(clause.Locking{Strength: "UPDATE"})
		}
		var receiver model.User
		if err := receiverQuery.First(&receiver, toUserID).Error; err != nil {
			return errors.New("penerima tidak ditemukan")
		}

		if err := tx.Model(&sender).Update("balance", sender.Balance-amount).Error; err != nil {
			return err
		}

		if err := tx.Model(&receiver).Update("balance", receiver.Balance+amount).Error; err != nil {
			return err
		}

		if err := s.txRepo.Record(tx, &model.Transaction{
			UserID:      fromUserID,
			Type:        "transfer_out",
			Amount:      amount,
			Description: senderNote,
		}); err != nil {
			return err
		}

		if err := s.txRepo.Record(tx, &model.Transaction{
			UserID:      toUserID,
			Type:        "transfer_in",
			Amount:      amount,
			Description: receiverNote,
		}); err != nil {
			return err
		}

		return nil
	})
}

func (s *walletService) GetHistory(userID uint, txType string, search string, page int, limit int) ([]model.Transaction, int64, error) {
	return s.txRepo.FindByUserID(userID, txType, search, page, limit)
}