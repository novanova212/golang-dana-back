package service

// Layer "service" ini isinya LOGIKA BISNIS.
// Contoh: "email harus unik", "password harus di-hash sebelum disimpan",
// "kalau login berhasil, buatkan token JWT".
//
// Analoginya di Laravel: ini seperti Service class kalau kamu memisahkan
// logika dari Controller (bukan taruh semua logic langsung di Controller).

import (
	"errors"
	"time"

	"dana-clone/internal/model"
	"dana-clone/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Interface: kontrak apa saja yang bisa dilakukan AuthService.
type AuthService interface {
	Register(name, email, password string) (*model.User, error)
	Login(email, password string) (string, error) // mengembalikan JWT token
}

type authService struct {
	userRepo  repository.UserRepository
	jwtSecret string
}

func NewAuthService(userRepo repository.UserRepository, jwtSecret string) AuthService {
	return &authService{userRepo: userRepo, jwtSecret: jwtSecret}
}

// Register mendaftarkan user baru.
func (s *authService) Register(name, email, password string) (*model.User, error) {
	// Cek dulu apakah email sudah dipakai. Mirip validasi 'unique:users,email'
	// di Laravel Form Request.
	existing, _ := s.userRepo.FindByEmail(email)
	if existing != nil {
		return nil, errors.New("email sudah terdaftar")
	}

	// Hash password sebelum disimpan. JANGAN PERNAH simpan password mentah.
	// bcrypt.GenerateFromPassword ini setara dengan Hash::make() di Laravel.
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Name:     name,
		Email:    email,
		Password: string(hashedPassword),
		Balance:  0,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

// Login memverifikasi email+password, lalu mengembalikan JWT token kalau benar.
func (s *authService) Login(email, password string) (string, error) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return "", errors.New("email atau password salah")
	}

	// bcrypt.CompareHashAndPassword ini setara dengan Hash::check() di Laravel.
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", errors.New("email atau password salah")
	}

	// Generate JWT token, mirip fungsi yang dilakukan Laravel Sanctum/Passport
	// saat createToken(), cuma di sini kita tulis manual.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(24 * time.Hour).Unix(), // token berlaku 24 jam
	})

	signedToken, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", err
	}

	return signedToken, nil
}
