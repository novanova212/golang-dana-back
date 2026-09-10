package handler

// Layer "handler" ini tugasnya:
// 1. Terima request dari HTTP (via Gin)
// 2. Ambil & validasi input dasar (format JSON, dst)
// 3. Panggil service untuk proses logika bisnis
// 4. Kembalikan response
//
// Ini PERSIS peran Controller di Laravel. Bedanya, di sini kita
// pakai *gin.Context (c) untuk baca request & kirim response,
// mirip $request dan return response()->json() di Laravel.

import (
	"net/http"

	"dana-clone/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Struct ini mendefinisikan bentuk JSON yang diharapkan masuk saat register.
// Mirip Form Request di Laravel (RegisterRequest dengan rules()).
// Tag `binding:"required"` artinya Gin akan otomatis menolak request
// kalau field ini kosong -  mirip 'required' di rules Laravel.
type RegisterInput struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// Register menangani POST /api/register
func (h *AuthHandler) Register(c *gin.Context) {
	var input RegisterInput

	// ShouldBindJSON otomatis parse body JSON ke struct di atas,
	// SEKALIGUS validasi berdasarkan tag `binding:"..."`.
	// Mirip $request->validate([...]) di Laravel, tapi jadi satu baris.
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.authService.Register(input.Name, input.Email, input.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// gin.H{...} itu shortcut untuk map[string]interface{},
	// dipakai untuk bikin response JSON. Mirip return response()->json([...]).
	c.JSON(http.StatusCreated, gin.H{
		"message": "Registrasi berhasil",
		"user":    user,
	})
}

// Login menangani POST /api/login
func (h *AuthHandler) Login(c *gin.Context) {
	var input LoginInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.authService.Login(input.Email, input.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login berhasil",
		"token":   token,
	})
}
