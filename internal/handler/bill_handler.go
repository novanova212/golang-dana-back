package handler

import (

	"fmt"
	"net/http"

	"dana-clone/internal/service"

	"github.com/gin-gonic/gin"
)

type BillHandler struct {
	billService service.BillService
}

func NewBillHandler(billService service.BillService) *BillHandler {
	return &BillHandler{billService: billService}
}

type CreateBillInput struct {
	CreatorID      uint   `json:"creator_id" binding:"required"`
	Title          string `json:"title" binding:"required"`
	TotalAmount    int64  `json:"total_amount" binding:"required"`
	ParticipantIDs []uint `json:"participant_ids" binding:"required"`
}

// CreateBill menangani POST /api/bills
func (h *BillHandler) CreateBill(c *gin.Context) {
	var input CreateBillInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	bill, err := h.billService.CreateBill(input.CreatorID, input.Title, input.TotalAmount, input.ParticipantIDs)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Bill berhasil dibuat",
		"bill":    bill,
	})
}

// GetBillDetail menangani GET /api/bills/:id
// ":id" ini parameter dinamis di URL, mirip {id} di route Laravel.
func (h *BillHandler) GetBillDetail(c *gin.Context) {
	// c.Param("id") mengambil nilai ":id" dari URL, misal /api/bills/5 -> "5"
	id := c.Param("id")

	var billID uint
	if _, err := fmt.Sscanf(id, "%d", &billID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID bill tidak valid"})
		return
	}

	bill, participants, err := h.billService.GetBillDetail(billID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"bill":         bill,
		"participants": participants,
	})
}

type SettleInput struct {
	UserID uint `json:"user_id" binding:"required"`
}

// SettleParticipant menangani POST /api/bills/participants/:participant_id/settle
func (h *BillHandler) SettleParticipant(c *gin.Context) {
	id := c.Param("participant_id")

	var participantID uint
	if _, err := fmt.Sscanf(id, "%d", &participantID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID peserta tidak valid"})
		return
	}

	var input SettleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.billService.SettleParticipant(participantID, input.UserID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Pembayaran berhasil, tagihan lunas"})
}