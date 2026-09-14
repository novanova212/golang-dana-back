package model

import "time"

// MoneyRequest mewakili permintaan uang dari satu user ke user lain.
// Requester = yang MEMINTA/MENAGIH uang.
// Target    = yang DIMINTA untuk membayar.
type MoneyRequest struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	RequesterID uint      `json:"requester_id" gorm:"not null"`
	Requester   User      `json:"requester" gorm:"foreignKey:RequesterID"`
	TargetID    uint      `json:"target_id" gorm:"not null"`
	Target      User      `json:"target" gorm:"foreignKey:TargetID"`
	Amount      int64     `json:"amount" gorm:"not null"`
	Description string    `json:"description"`
	Status      string    `json:"status" gorm:"default:pending"` // pending, paid, declined
	CreatedAt   time.Time `json:"created_at"`
}