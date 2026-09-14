package service

import (
	"errors"

	"dana-clone/internal/model"
	"dana-clone/internal/repository"
)

// ParticipantShare mewakili satu peserta dan porsi custom yang harus dia bayar.
type ParticipantShare struct {
	UserID uint
	Amount int64
}

type BillService interface {
	CreateBill(creatorID uint, title string, totalAmount int64, participantIDs []uint) (*model.Bill, error)
	CreateCustomBill(creatorID uint, title string, totalAmount int64, shares []ParticipantShare) (*model.Bill, error)
	GetBillDetail(billID uint) (*model.Bill, []model.BillParticipant, error)
	SettleParticipant(participantID uint, payingUserID uint) error
}

type billService struct {
	billRepo      repository.BillRepository
	walletService WalletService
}

func NewBillService(billRepo repository.BillRepository, walletService WalletService) BillService {
	return &billService{billRepo: billRepo, walletService: walletService}
}

// CreateBill: cara LAMA, bagi rata semua orang. Tetap dipertahankan
// supaya fitur yang sudah ada sebelumnya tidak rusak.
func (s *billService) CreateBill(creatorID uint, title string, totalAmount int64, participantIDs []uint) (*model.Bill, error) {
	if totalAmount <= 0 {
		return nil, errors.New("total tagihan harus lebih dari 0")
	}
	if len(participantIDs) == 0 {
		return nil, errors.New("minimal ada 1 peserta selain kamu")
	}
	if err := validateParticipantIDs(creatorID, participantIDs); err != nil {
		return nil, err
	}

	numPeople := int64(len(participantIDs) + 1)
	share := totalAmount / numPeople

	shares := make([]ParticipantShare, 0, len(participantIDs))
	for _, uid := range participantIDs {
		shares = append(shares, ParticipantShare{UserID: uid, Amount: share})
	}

	return s.createBillInternal(creatorID, title, totalAmount, shares)
}

// CreateCustomBill: cara BARU, tiap peserta bisa punya porsi berbeda-beda.
// Contoh: Budi (creator) bayar duluan 300rb, Ani nanggung 150rb, Citra 100rb
// (sisanya 50rb otomatis jadi porsi Budi sendiri, karena dia creator).
func (s *billService) CreateCustomBill(creatorID uint, title string, totalAmount int64, shares []ParticipantShare) (*model.Bill, error) {
	if totalAmount <= 0 {
		return nil, errors.New("total tagihan harus lebih dari 0")
	}
	if len(shares) == 0 {
		return nil, errors.New("minimal ada 1 peserta selain kamu")
	}

	ids := make([]uint, 0, len(shares))
	for _, sh := range shares {
		ids = append(ids, sh.UserID)
	}
	if err := validateParticipantIDs(creatorID, ids); err != nil {
		return nil, err
	}

	// Jumlahkan semua porsi peserta (TIDAK termasuk creator).
	var totalShares int64
	for _, s := range shares {
		if s.Amount <= 0 {
			return nil, errors.New("porsi tiap peserta harus lebih dari 0")
		}
		totalShares += s.Amount
	}

	// Validasi: total porsi peserta TIDAK BOLEH melebihi total tagihan,
	// karena sisanya (total - totalShares) otomatis jadi porsi creator sendiri.
	if totalShares > totalAmount {
		return nil, errors.New("total porsi peserta melebihi total tagihan")
	}

	return s.createBillInternal(creatorID, title, totalAmount, shares)
}

// validateParticipantIDs memastikan: (1) tidak ada peserta yang ID-nya
// sama dengan creator sendiri (creator tidak "berutang" ke dirinya sendiri),
// dan (2) tidak ada ID yang muncul dua kali (biar tidak dobel dicatat).
func validateParticipantIDs(creatorID uint, ids []uint) error {
	seen := make(map[uint]bool)
	for _, id := range ids {
		if id == creatorID {
			return errors.New("creator tidak boleh menjadi peserta juga")
		}
		if seen[id] {
			return errors.New("ada user ID yang duplikat di daftar peserta")
		}
		seen[id] = true
	}
	return nil
}

// createBillInternal adalah logic bersama yang dipakai baik oleh
// CreateBill (rata) maupun CreateCustomBill (custom) - menghindari
// duplikasi kode simpan-ke-database.
func (s *billService) createBillInternal(creatorID uint, title string, totalAmount int64, shares []ParticipantShare) (*model.Bill, error) {
	bill := &model.Bill{
		Title:       title,
		TotalAmount: totalAmount,
		CreatorID:   creatorID,
	}
	if err := s.billRepo.CreateBill(bill); err != nil {
		return nil, err
	}

	participants := make([]model.BillParticipant, 0, len(shares))
	for _, share := range shares {
		participants = append(participants, model.BillParticipant{
			BillID: bill.ID,
			UserID: share.UserID,
			Amount: share.Amount,
			Paid:   false,
		})
	}

	if err := s.billRepo.CreateParticipants(participants); err != nil {
		return nil, err
	}

	return bill, nil
}

func (s *billService) GetBillDetail(billID uint) (*model.Bill, []model.BillParticipant, error) {
	bill, err := s.billRepo.FindBillByID(billID)
	if err != nil {
		return nil, nil, errors.New("bill tidak ditemukan")
	}

	participants, err := s.billRepo.FindParticipantsByBillID(billID)
	if err != nil {
		return nil, nil, err
	}

	return bill, participants, nil
}

// CalculateCreatorShare menghitung porsi yang ditanggung creator sendiri:
// total tagihan dikurangi jumlah semua porsi peserta lain.
func CalculateCreatorShare(bill *model.Bill, participants []model.BillParticipant) int64 {
	var totalParticipantShares int64
	for _, p := range participants {
		totalParticipantShares += p.Amount
	}
	return bill.TotalAmount - totalParticipantShares
}

func (s *billService) SettleParticipant(participantID uint, payingUserID uint) error {
	participant, err := s.billRepo.FindParticipantByID(participantID)
	if err != nil {
		return errors.New("data peserta tidak ditemukan")
	}

	if participant.UserID != payingUserID {
		return errors.New("kamu tidak berhak melunasi tagihan ini")
	}

	if participant.Paid {
		return errors.New("tagihan ini sudah lunas")
	}

	bill, err := s.billRepo.FindBillByID(participant.BillID)
	if err != nil {
		return errors.New("bill tidak ditemukan")
	}

	if err := s.walletService.Transfer(payingUserID, bill.CreatorID, participant.Amount); err != nil {
		return err
	}

	return s.billRepo.MarkParticipantPaid(participantID)
}