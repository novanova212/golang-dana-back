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
	TargetID    uint   `json:"target_id" binding:"required"`
	Amount      int64  `json:"amount" binding:"required"`
	Description string `json:"description"`
}

// CreateRequest menangani POST /api/requests (PROTECTED).
// requester_id diambil dari token - mencegah orang membuat tagihan
// "atas nama" orang lain ke pihak ketiga.
func (h *MoneyRequestHandler) CreateRequest(c *gin.Context) {
	requesterID, ok := getUserID(c)
	if !ok {
		return
	}

	var input CreateRequestInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req, err := h.reqService.CreateRequest(requesterID, input.TargetID, input.Amount, input.Description)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Permintaan berhasil dibuat", "request": req})
}

// GetIncoming menangani GET /api/requests/incoming (PROTECTED, tanpa param).
func (h *MoneyRequestHandler) GetIncoming(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}
	reqs, err := h.reqService.GetIncoming(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"requests": reqs})
}

// GetOutgoing menangani GET /api/requests/outgoing (PROTECTED, tanpa param).
func (h *MoneyRequestHandler) GetOutgoing(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}
	reqs, err := h.reqService.GetOutgoing(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"requests": reqs})
}

func (h *MoneyRequestHandler) PayRequest(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}
	var requestID uint
	if _, err := fmt.Sscanf(c.Param("id"), "%d", &requestID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}
	if err := h.reqService.PayRequest(requestID, userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Pembayaran berhasil"})
}

func (h *MoneyRequestHandler) DeclineRequest(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}
	var requestID uint
	if _, err := fmt.Sscanf(c.Param("id"), "%d", &requestID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}
	if err := h.reqService.DeclineRequest(requestID, userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Permintaan ditolak"})
}