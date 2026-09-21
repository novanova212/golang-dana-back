package service

// Test ini pakai FAKE (test double) untuk BillRepository dan WalletService,
// BUKAN database beneran. Ini pola umum di unit testing: kita hanya ingin
// menguji LOGIC di bill_service.go secara terisolasi, tanpa bergantung
// pada database atau service lain yang sudah punya test-nya sendiri.
//
// Jalankan dengan: go test ./internal/service/... -v
// Perlu: go get github.com/stretchr/testify

import (
	"errors"
	"testing"

	"dana-clone/internal/model"

	"github.com/stretchr/testify/assert"
)

// ---- Fake BillRepository ----

type fakeBillRepository struct {
	bills        map[uint]*model.Bill
	participants map[uint][]model.BillParticipant
	nextBillID   uint
	nextPartID   uint
}

func newFakeBillRepository() *fakeBillRepository {
	return &fakeBillRepository{
		bills:        make(map[uint]*model.Bill),
		participants: make(map[uint][]model.BillParticipant),
		nextBillID:   1,
		nextPartID:   1,
	}
}

func (f *fakeBillRepository) CreateBill(bill *model.Bill) error {
	bill.ID = f.nextBillID
	f.nextBillID++
	f.bills[bill.ID] = bill
	return nil
}

func (f *fakeBillRepository) CreateParticipants(participants []model.BillParticipant) error {
	for i := range participants {
		participants[i].ID = f.nextPartID
		f.nextPartID++
		f.participants[participants[i].BillID] = append(f.participants[participants[i].BillID], participants[i])
	}
	return nil
}

func (f *fakeBillRepository) FindBillByID(id uint) (*model.Bill, error) {
	bill, ok := f.bills[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return bill, nil
}

func (f *fakeBillRepository) FindParticipantByID(id uint) (*model.BillParticipant, error) {
	for _, list := range f.participants {
		for i := range list {
			if list[i].ID == id {
				return &list[i], nil
			}
		}
	}
	return nil, errors.New("not found")
}

func (f *fakeBillRepository) MarkParticipantPaid(participantID uint) error {
	for billID, list := range f.participants {
		for i := range list {
			if list[i].ID == participantID {
				f.participants[billID][i].Paid = true
				return nil
			}
		}
	}
	return errors.New("not found")
}

func (f *fakeBillRepository) FindParticipantsByBillID(billID uint) ([]model.BillParticipant, error) {
	return f.participants[billID], nil
}

// ---- Fake WalletService ----
// Mencatat pemanggilan Transfer, dan bisa diatur untuk gagal (simulasi
// saldo tidak cukup) lewat field failTransfer.

type fakeWalletService struct {
	failTransfer  bool
	transferCalls int
}

func (f *fakeWalletService) TopUp(userID uint, amount int64) (int64, error) { return 0, nil }
func (f *fakeWalletService) Transfer(fromUserID, toUserID uint, amount int64) error {
	return f.TransferWithNote(fromUserID, toUserID, amount, "", "")
}
func (f *fakeWalletService) TransferWithNote(fromUserID, toUserID uint, amount int64, senderNote, receiverNote string) error {
	f.transferCalls++
	if f.failTransfer {
		return errors.New("saldo tidak cukup")
	}
	return nil
}
func (f *fakeWalletService) GetHistory(userID uint, txType, search string, page, limit int) ([]model.Transaction, int64, error) {
	return nil, 0, nil
}

// ---- Tests ----

func TestCreateBill_BagiRata(t *testing.T) {
	repo := newFakeBillRepository()
	wallet := &fakeWalletService{}
	svc := NewBillService(repo, wallet)

	bill, err := svc.CreateBill(1, "Makan malam", 300000, []uint{2, 3})

	assert.NoError(t, err)
	assert.Equal(t, int64(300000), bill.TotalAmount)

	participants, _ := repo.FindParticipantsByBillID(bill.ID)
	assert.Len(t, participants, 2)
	// 300000 / 3 orang (creator + 2 participant) = 100000 masing-masing
	assert.Equal(t, int64(100000), participants[0].Amount)
	assert.Equal(t, int64(100000), participants[1].Amount)
}

func TestCreateBill_TotalAmountInvalid(t *testing.T) {
	repo := newFakeBillRepository()
	wallet := &fakeWalletService{}
	svc := NewBillService(repo, wallet)

	_, err := svc.CreateBill(1, "Test", 0, []uint{2})
	assert.Error(t, err)
}

func TestCreateCustomBill_Sukses(t *testing.T) {
	repo := newFakeBillRepository()
	wallet := &fakeWalletService{}
	svc := NewBillService(repo, wallet)

	shares := []ParticipantShare{
		{UserID: 2, Amount: 150000},
		{UserID: 3, Amount: 100000},
	}
	bill, err := svc.CreateCustomBill(1, "Custom", 300000, shares)

	assert.NoError(t, err)
	participants, _ := repo.FindParticipantsByBillID(bill.ID)
	assert.Len(t, participants, 2)
}

func TestCreateCustomBill_TotalMelebihi(t *testing.T) {
	repo := newFakeBillRepository()
	wallet := &fakeWalletService{}
	svc := NewBillService(repo, wallet)

	shares := []ParticipantShare{
		{UserID: 2, Amount: 200000},
		{UserID: 3, Amount: 200000},
	}
	_, err := svc.CreateCustomBill(1, "Custom", 300000, shares)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "melebihi")
}

func TestCreateCustomBill_CreatorTidakBolehJadiPeserta(t *testing.T) {
	repo := newFakeBillRepository()
	wallet := &fakeWalletService{}
	svc := NewBillService(repo, wallet)

	shares := []ParticipantShare{
		{UserID: 1, Amount: 100000}, // creator sendiri = 1, sama seperti pemanggil
	}
	_, err := svc.CreateCustomBill(1, "Custom", 300000, shares)

	assert.Error(t, err)
}

func TestCreateCustomBill_UserIDDuplikat(t *testing.T) {
	repo := newFakeBillRepository()
	wallet := &fakeWalletService{}
	svc := NewBillService(repo, wallet)

	shares := []ParticipantShare{
		{UserID: 2, Amount: 50000},
		{UserID: 2, Amount: 50000},
	}
	_, err := svc.CreateCustomBill(1, "Custom", 300000, shares)

	assert.Error(t, err)
}

func TestSettleParticipant_Sukses(t *testing.T) {
	repo := newFakeBillRepository()
	wallet := &fakeWalletService{}
	svc := NewBillService(repo, wallet)

	bill, _ := svc.CreateBill(1, "Test", 200000, []uint{2})
	participants, _ := repo.FindParticipantsByBillID(bill.ID)
	participantID := participants[0].ID

	err := svc.SettleParticipant(participantID, 2)

	assert.NoError(t, err)
	assert.Equal(t, 1, wallet.transferCalls)

	updated, _ := repo.FindParticipantByID(participantID)
	assert.True(t, updated.Paid)
}

func TestSettleParticipant_BukanPemilik(t *testing.T) {
	repo := newFakeBillRepository()
	wallet := &fakeWalletService{}
	svc := NewBillService(repo, wallet)

	bill, _ := svc.CreateBill(1, "Test", 200000, []uint{2})
	participants, _ := repo.FindParticipantsByBillID(bill.ID)
	participantID := participants[0].ID

	// User 999 mencoba melunasi utang milik user 2 - harus ditolak.
	err := svc.SettleParticipant(participantID, 999)

	assert.Error(t, err)
	assert.Equal(t, 0, wallet.transferCalls)
}

func TestSettleParticipant_SaldoTidakCukup(t *testing.T) {
	repo := newFakeBillRepository()
	wallet := &fakeWalletService{failTransfer: true}
	svc := NewBillService(repo, wallet)

	bill, _ := svc.CreateBill(1, "Test", 200000, []uint{2})
	participants, _ := repo.FindParticipantsByBillID(bill.ID)
	participantID := participants[0].ID

	err := svc.SettleParticipant(participantID, 2)

	assert.Error(t, err)
	updated, _ := repo.FindParticipantByID(participantID)
	assert.False(t, updated.Paid, "participant tidak boleh ditandai lunas kalau transfer gagal")
}