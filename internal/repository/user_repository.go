package repository

// Layer "repository" ini tugasnya CUMA satu: bicara dengan database.
// Dia tidak tahu-menahu soal logika bisnis (misal validasi password),
// itu urusan layer "service" nanti.
//
// Analoginya di Laravel: kalau kamu pernah pisahkan query dari Controller
// ke class Repository (bukan langsung User::where(...) di Controller),
// ini persis konsep yang sama.

import (
	"dana-clone/internal/model"

	"gorm.io/gorm"
)

// Interface ini mendefinisikan "kontrak": apa saja yang BISA dilakukan
// repository ini, tanpa peduli bagaimana caranya.
// Gunanya: nanti kalau mau bikin versi palsu (mock) untuk testing,
// atau ganti dari PostgreSQL ke database lain, kita tinggal buat
// implementasi baru dari interface ini tanpa mengubah kode service.
type UserRepository interface {
	Create(user *model.User) error
	FindByEmail(email string) (*model.User, error)
	FindByID(id uint) (*model.User, error)
	UpdateBalance(userID uint, newBalance int64) error
}

// Struct ini adalah implementasi NYATA dari interface di atas,
// menggunakan GORM + PostgreSQL.
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository adalah "constructor" - fungsi untuk membuat instance
// userRepository baru. Pola ini sering disebut dependency injection manual:
// kita "suntik" koneksi db ke dalam repository saat dibuat.
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

// Create menyimpan user baru ke database.
// Mirip User::create([...]) di Eloquent.
func (r *userRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

// FindByEmail mencari satu user berdasarkan email.
// Mirip User::where('email', $email)->first() di Eloquent.
func (r *userRepository) FindByEmail(email string) (*model.User, error) {
	var user model.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByID mencari user berdasarkan ID.
// Mirip User::find($id) di Eloquent.
func (r *userRepository) FindByID(id uint) (*model.User, error) {
	var user model.User
	err := r.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateBalance mengubah saldo user.
// Nanti dipakai untuk fitur top up & transfer.
func (r *userRepository) UpdateBalance(userID uint, newBalance int64) error {
	return r.db.Model(&model.User{}).Where("id = ?", userID).Update("balance", newBalance).Error
}
