package service

import (
	"errors"
	"time"

	"dana-clone/internal/model"
	"dana-clone/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Register(name, email, password string) (*model.User, error)
	Login(email, password string) (string, error)
	UpdateProfile(userID uint, name string) error
	ChangePassword(userID uint, oldPassword, newPassword string) error
}

type authService struct {
	userRepo  repository.UserRepository
	jwtSecret string
}

func NewAuthService(userRepo repository.UserRepository, jwtSecret string) AuthService {
	return &authService{userRepo: userRepo, jwtSecret: jwtSecret}
}

func (s *authService) Register(name, email, password string) (*model.User, error) {
	existing, _ := s.userRepo.FindByEmail(email)
	if existing != nil {
		return nil, errors.New("email sudah terdaftar")
	}

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

func (s *authService) Login(email, password string) (string, error) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return "", errors.New("email atau password salah")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", errors.New("email atau password salah")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})

	signedToken, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

// UpdateProfile mengubah nama user. userID diambil dari token JWT
// (lewat middleware), BUKAN dari input body - supaya orang lain tidak
// bisa mengubah profil orang lain hanya dengan menebak user_id.
func (s *authService) UpdateProfile(userID uint, name string) error {
	if name == "" {
		return errors.New("nama tidak boleh kosong")
	}
	return s.userRepo.UpdateName(userID, name)
}

// ChangePassword mengganti password user, mewajibkan password LAMA
// yang benar terlebih dahulu (supaya orang yang sekadar tahu token
// bocor tidak otomatis bisa ganti password tanpa tahu password asli).
func (s *authService) ChangePassword(userID uint, oldPassword, newPassword string) error {
	if len(newPassword) < 6 {
		return errors.New("password baru minimal 6 karakter")
	}

	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return errors.New("user tidak ditemukan")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return errors.New("password lama salah")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.userRepo.UpdatePassword(userID, string(hashedPassword))
}