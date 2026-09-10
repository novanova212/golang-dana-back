package service

import (
	"errors"

	"dana-clone/internal/model"
	"dana-clone/internal/repository"
)

type BillService interface {
	CreateBill(creatorID uint, title string, totalAmount int64, participantIDs []uint) (*model.Bill, error)
	GetBillDetail(billID uint) (*model.Bill, []model.BillParticipant, error)
	SettleParticipant(participantID uint, payingUserID uint) error
}

type billService struct {
	billRepo      repository.BillRepository
	walletService WalletService // <- pakai ulang service transfer yang sudah ada!
}

func NewBillService(billRepo repository.BillRepository, walletService WalletService) BillService {
	return &billService{billRepo: billRepo, walletService: walletService}
}

// CreateBill membuat tagihan baru dan membagi rata ke semua peserta
// (termasuk creator sendiri, tapi porsi creator tidak dicatat sebagai utang
// karena dialah yang bayar duluan).
func (s *billService) CreateBill(creatorID uint, title string, totalAmount int64, participantIDs []uint) (*model.Bill, error) {
	if totalAmount <= 0 {
		return nil, errors.New("total tagihan harus lebih dari 0")
	}
	if len(participantIDs) == 0 {
		return nil, errors.New("minimal ada 1 peserta selain kamu")
	}

	// Jumlah orang yang menanggung = creator + semua participant.
	numPeople := int64(len(participantIDs) + 1)
	share := totalAmount / numPeople
	// Catatan: pembagian ini pakai integer division, jadi kalau ada sisa
	// (misal 100.000 dibagi 3 = 33.333,33), sisanya akan hilang/terpotong.
	// Untuk tahap belajar ini kita abaikan dulu, nanti bisa kita
	// sempurnakan (misal sisa dibebankan ke creator).

	bill := &model.Bill{
		Title:       title,
		TotalAmount: totalAmount,
		CreatorID:   creatorID,
	}
	if err := s.billRepo.CreateBill(bill); err != nil {
		return nil, err
	}

	// Bikin baris BillParticipant untuk tiap peserta (BUKAN untuk creator,
	// karena creator tidak "berutang" ke dirinya sendiri).
	participants := make([]model.BillParticipant, 0, len(participantIDs))
	for _, uid := range participantIDs {
		participants = append(participants, model.BillParticipant{
			BillID: bill.ID,
			UserID: uid,
			Amount: share,
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

// SettleParticipant dipanggil ketika seorang peserta mau "melunasi" utangnya.
// Ini akan memicu Transfer BENERAN dari peserta ke creator bill.
func (s *billService) SettleParticipant(participantID uint, payingUserID uint) error {
	participant, err := s.billRepo.FindParticipantByID(participantID)
	if err != nil {
		return errors.New("data peserta tidak ditemukan")
	}

	// Pastikan yang mau bayar adalah ORANG YANG MEMANG BERUTANG di baris ini,
	// bukan orang lain yang asal panggil endpoint ini.
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

	// Pakai ulang logic Transfer yang SUDAH TERUJI dari fitur sebelumnya.
	// Ini keuntungan besar dari clean architecture: kita tidak perlu
	// menulis ulang logic "pindah saldo dengan aman", cukup panggil
	// service yang sudah ada.
	if err := s.walletService.Transfer(payingUserID, bill.CreatorID, participant.Amount); err != nil {
		return err
	}

	return s.billRepo.MarkParticipantPaid(participantID)
}