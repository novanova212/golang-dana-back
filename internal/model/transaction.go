package model

import "time"

// Transaction mencatat SETIAP pergerakan saldo yang terjadi:
// top up, transfer keluar, transfer masuk, atau settle bill.
// Mirip tabel "audit log" atau "mutasi" di aplikasi bank/e-wallet beneran.
type Transaction struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	UserID      uint      `json:"user_id" gorm:"not null;index"` // milik siapa baris riwayat ini
	Type        string    `json:"type" gorm:"not null"`          // "topup", "transfer_out", "transfer_in"
	Amount      int64     `json:"amount" gorm:"not null"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}