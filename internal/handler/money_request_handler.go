package handler

import (
	"fmt"
	"net/http"

	"dana-clone/internal/service"

	"github.com/gin-gonic/gin"
)

type MoneyRequestHandler struct {
	reqService service.MoneyRequestService
}

func NewMoneyRequestHandler(reqService service.MoneyRequestService) *MoneyRequestHandler {
	return &MoneyRequestHandler{reqService: reqService}
}

type CreateRequestInput struct {
	RequesterID uint   `json:"requester_id" binding:"required"`
	TargetID    uint   `json:"target_id" binding:"required"`
	Amount      int64  `json:"amount" binding:"required"`
	Description string `json:"description"`
}

// CreateRequest menangani POST /api/requests
func (h *MoneyRequestHandler) CreateRequest(c *gin.Context) {
	var input CreateRequestInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req, err := h.reqService.CreateRequest(input.RequesterID, input.TargetID, input.Amount, input.Description)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Permintaan berhasil dibuat", "request": req})
}

// GetIncoming menangani GET /api/requests/incoming/:user_id
func (h *MoneyRequestHandler) GetIncoming(c *gin.Context) {
	var userID uint
	if _, err := fmt.Sscanf(c.Param("user_id"), "%d", &userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID tidak valid"})
		return
	}
	reqs, err := h.reqService.GetIncoming(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"requests": reqs})
}

// GetOutgoing menangani GET /api/requests/outgoing/:user_id
func (h *MoneyRequestHandler) GetOutgoing(c *gin.Context) {
	var userID uint
	if _, err := fmt.Sscanf(c.Param("user_id"), "%d", &userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID tidak valid"})
		return
	}
	reqs, err := h.reqService.GetOutgoing(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"requests": reqs})
}

type RequestActionInput struct {
	UserID uint `json:"user_id" binding:"required"`
}

// PayRequest menangani POST /api/requests/:id/pay
func (h *MoneyRequestHandler) PayRequest(c *gin.Context) {
	var requestID uint
	if _, err := fmt.Sscanf(c.Param("id"), "%d", &requestID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}
	var input RequestActionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.reqService.PayRequest(requestID, input.UserID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Pembayaran berhasil"})
}

// DeclineRequest menangani POST /api/requests/:id/decline
func (h *MoneyRequestHandler) DeclineRequest(c *gin.Context) {
	var requestID uint
	if _, err := fmt.Sscanf(c.Param("id"), "%d", &requestID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}
	var input RequestActionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.reqService.DeclineRequest(requestID, input.UserID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Permintaan ditolak"})
}