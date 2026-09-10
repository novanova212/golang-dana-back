package repository

import (
	"dana-clone/internal/model"

	"gorm.io/gorm"
)

type BillRepository interface {
	CreateBill(bill *model.Bill) error
	CreateParticipants(participants []model.BillParticipant) error
	FindBillByID(id uint) (*model.Bill, error)
	FindParticipantByID(id uint) (*model.BillParticipant, error)
	MarkParticipantPaid(participantID uint) error
	FindParticipantsByBillID(billID uint) ([]model.BillParticipant, error)
}

type billRepository struct {
	db *gorm.DB
}

func NewBillRepository(db *gorm.DB) BillRepository {
	return &billRepository{db: db}
}

func (r *billRepository) CreateBill(bill *model.Bill) error {
	return r.db.Create(bill).Error
}

// CreateParticipants menyimpan banyak baris sekaligus.
// Mirip Model::insert([...]) kalau di Eloquent untuk bulk insert.
func (r *billRepository) CreateParticipants(participants []model.BillParticipant) error {
	return r.db.Create(&participants).Error
}

func (r *billRepository) FindBillByID(id uint) (*model.Bill, error) {
	var bill model.Bill
	if err := r.db.Preload("Creator").First(&bill, id).Error; err != nil {
		return nil, err
	}
	return &bill, nil
}

func (r *billRepository) FindParticipantByID(id uint) (*model.BillParticipant, error) {
	var p model.BillParticipant
	if err := r.db.First(&p, id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *billRepository) MarkParticipantPaid(participantID uint) error {
	return r.db.Model(&model.BillParticipant{}).Where("id = ?", participantID).Update("paid", true).Error
}

// FindParticipantsByBillID mengambil semua peserta dalam satu bill,
// sekalian "Preload" data User-nya (mirip with('user') di Eloquent).
func (r *billRepository) FindParticipantsByBillID(billID uint) ([]model.BillParticipant, error) {
	var participants []model.BillParticipant
	if err := r.db.Preload("User").Where("bill_id = ?", billID).Find(&participants).Error; err != nil {
		return nil, err
	}
	return participants, nil
}