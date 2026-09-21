package handler

import (
	"net/http"
	"strconv"

	"dana-clone/internal/service"

	"github.com/gin-gonic/gin"
)

type WalletHandler struct {
	walletService service.WalletService
}

func NewWalletHandler(walletService service.WalletService) *WalletHandler {
	return &WalletHandler{walletService: walletService}
}

// getUserID adalah helper untuk mengambil user_id dari context yang sudah
// dititipkan AuthMiddleware. Dipakai di semua handler protected di file ini
// supaya konsisten dan tidak duplikasi kode.
func getUserID(c *gin.Context) (uint, bool) {
	val, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tidak terautentikasi"})
		return 0, false
	}
	return val.(uint), true
}

type TopUpInput struct {
	Amount int64 `json:"amount" binding:"required"`
}

// TopUp menangani POST /api/wallet/topup (PROTECTED).
// user_id diambil dari token, BUKAN dari body - mencegah orang lain
// top up saldo atas nama akun orang lain.
func (h *WalletHandler) TopUp(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	var input TopUpInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newBalance, err := h.walletService.TopUp(userID, input.Amount)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Top up berhasil", "new_balance": newBalance})
}

type TransferInput struct {
	ToUserID uint  `json:"to_user_id" binding:"required"`
	Amount   int64 `json:"amount" binding:"required"`
}

// Transfer menangani POST /api/wallet/transfer (PROTECTED).
// from_user_id diambil dari token - inilah perbaikan keamanan utamanya:
// dulu siapa saja bisa transfer "atas nama" user lain hanya dengan tahu
// ID-nya, karena from_user_id dikirim bebas lewat body.
func (h *WalletHandler) Transfer(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	var input TransferInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.walletService.Transfer(userID, input.ToUserID, input.Amount); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Transfer berhasil"})
}

// GetHistory menangani GET /api/wallet/history (PROTECTED).
// Query params opsional: ?type=&search=&page=&limit=
func (h *WalletHandler) GetHistory(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	txType := c.Query("type")
	search := c.Query("search")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	history, total, err := h.walletService.GetHistory(userID, txType, search, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"history": history,
		"page":    page,
		"limit":   limit,
		"total":   total,
	})
}