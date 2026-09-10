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
}

type walletService struct {
	userRepo repository.UserRepository
	db       *gorm.DB
}

func NewWalletService(userRepo repository.UserRepository, db *gorm.DB) WalletService {
	return &walletService{userRepo: userRepo, db: db}
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

	return newBalance, nil
}

func (s *walletService) Transfer(fromUserID, toUserID uint, amount int64) error {
	if amount <= 0 {
		return errors.New("jumlah transfer harus lebih dari 0")
	}

	if fromUserID == toUserID {
		return errors.New("tidak bisa transfer ke diri sendiri")
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		var sender model.User
		// clause.Locking{Strength: "UPDATE"} ini cara yang benar di GORM v2
		// untuk mengunci baris (setara SELECT ... FOR UPDATE di SQL).
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

		return nil
	})
}