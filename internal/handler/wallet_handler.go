package handler

import (
	"net/http"

	"dana-clone/internal/service"

	"github.com/gin-gonic/gin"
)

type WalletHandler struct {
	walletService service.WalletService
}

func NewWalletHandler(walletService service.WalletService) *WalletHandler {
	return &WalletHandler{walletService: walletService}
}

type TopUpInput struct {
	UserID uint  `json:"user_id" binding:"required"`
	Amount int64 `json:"amount" binding:"required"`
}

func (h *WalletHandler) TopUp(c *gin.Context) {
	var input TopUpInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newBalance, err := h.walletService.TopUp(input.UserID, input.Amount)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Top up berhasil",
		"new_balance": newBalance,
	})
}

type TransferInput struct {
	FromUserID uint  `json:"from_user_id" binding:"required"`
	ToUserID   uint  `json:"to_user_id" binding:"required"`
	Amount     int64 `json:"amount" binding:"required"`
}

func (h *WalletHandler) Transfer(c *gin.Context) {
	var input TransferInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.walletService.Transfer(input.FromUserID, input.ToUserID, input.Amount); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Transfer berhasil",
	})
}