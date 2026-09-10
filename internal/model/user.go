package model

import "time"

// Struct ini mirip class User extends Model di Laravel/Eloquent.
// GORM akan otomatis bikin tabel "users" dari struct ini (lewat AutoMigrate).
//
// Tag `gorm:"..."` dan `json:"..."` di belakang tiap field:
// - json:"..."   -> nama field ini kalau di-convert ke JSON (response API)
// - gorm:"..."   -> aturan untuk kolom database (unique, not null, dst)
type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"not null"`
	Email     string    `json:"email" gorm:"unique;not null"`
	Password  string    `json:"-" gorm:"not null"` // json:"-" artinya field ini TIDAK PERNAH ikut ter-kirim di response API (biar password hash tidak bocor)
	Balance   int64     `json:"balance" gorm:"default:0"` // saldo disimpan dalam satuan terkecil (misal rupiah utuh, tanpa desimal)
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
