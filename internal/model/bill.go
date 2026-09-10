package model

import "time"

// Bill mewakili satu tagihan yang dibuat, misal "Makan malam bareng".
type Bill struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Title       string    `json:"title" gorm:"not null"`
	TotalAmount int64     `json:"total_amount" gorm:"not null"`
	CreatorID   uint      `json:"creator_id" gorm:"not null"` // user yang bayar duluan / yang berhak nagih
	Creator     User      `json:"creator" gorm:"foreignKey:CreatorID"`
	CreatedAt   time.Time `json:"created_at"`
}

// BillParticipant mewakili "porsi utang" satu orang di dalam sebuah bill.
// Satu Bill bisa punya banyak BillParticipant (relasi one-to-many),
// mirip relasi hasMany di Eloquent (Bill hasMany BillParticipant).
type BillParticipant struct {
	ID     uint  `json:"id" gorm:"primaryKey"`
	BillID uint  `json:"bill_id" gorm:"not null"`
	UserID uint  `json:"user_id" gorm:"not null"`
	User   User  `json:"user" gorm:"foreignKey:UserID"`
	Amount int64 `json:"amount" gorm:"not null"` // porsi yang harus dibayar orang ini
	Paid   bool  `json:"paid" gorm:"default:false"`
}