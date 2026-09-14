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
	// TransferWithNote sama seperti Transfer, tapi deskripsi riwayatnya
	// bisa dikustomisasi. Dipakai fitur lain (misal Split Bill) supaya
	// riwayat transaksinya jelas asal-usulnya, bukan cuma "Transfer keluar".
	TransferWithNote(fromUserID, toUserID uint, amount int64, senderNote, receiverNote string) error
	GetHistory(userID uint, txType string, search string) ([]model.Transaction, error)
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

	// Catat riwayatnya. 'nil' di sini artinya "pakai koneksi db biasa"
	// (top up bukan operasi 2 langkah, jadi tidak butuh database transaction).
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

// TransferWithNote adalah logic INTI perpindahan saldo (sama persis seperti
// sebelumnya: pakai database transaction + row locking supaya aman dari
// race condition), hanya saja deskripsi riwayatnya sekarang bisa diatur
// sesuai konteks pemanggilnya (transfer biasa, atau pelunasan split bill).
func (s *walletService) TransferWithNote(fromUserID, toUserID uint, amount int64, senderNote, receiverNote string) error {
	if amount <= 0 {
		return errors.New("jumlah transfer harus lebih dari 0")
	}

	if fromUserID == toUserID {
		return errors.New("tidak bisa transfer ke diri sendiri")
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		var sender model.User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&sender, fromUserID).Error; err != nil {
			return errors.New("pengirim tidak ditemukan")
		}

		if sender.Balance < amount {
			return errors.New("saldo tidak cukup")
		}

		var receiver model.User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&receiver, toUserID).Error; err != nil {
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

func (s *walletService) GetHistory(userID uint, txType string, search string) ([]model.Transaction, error) {
	return s.txRepo.FindByUserID(userID, txType, search)
}