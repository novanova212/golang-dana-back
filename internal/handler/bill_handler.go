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
	Title          string `json:"title" binding:"required"`
	TotalAmount    int64  `json:"total_amount" binding:"required"`
	ParticipantIDs []uint `json:"participant_ids" binding:"required"`
}

// CreateBill menangani POST /api/bills (PROTECTED).
// creator_id diambil dari token - mencegah orang membuat tagihan
// "atas nama" orang lain.
func (h *BillHandler) CreateBill(c *gin.Context) {
	creatorID, ok := getUserID(c)
	if !ok {
		return
	}

	var input CreateBillInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	bill, err := h.billService.CreateBill(creatorID, input.Title, input.TotalAmount, input.ParticipantIDs)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Bill berhasil dibuat", "bill": bill})
}

type ParticipantShareInput struct {
	UserID uint  `json:"user_id" binding:"required"`
	Amount int64 `json:"amount" binding:"required"`
}

type CreateCustomBillInput struct {
	Title        string                  `json:"title" binding:"required"`
	TotalAmount  int64                   `json:"total_amount" binding:"required"`
	Participants []ParticipantShareInput `json:"participants" binding:"required"`
}

func (h *BillHandler) CreateCustomBill(c *gin.Context) {
	creatorID, ok := getUserID(c)
	if !ok {
		return
	}

	var input CreateCustomBillInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	shares := make([]service.ParticipantShare, 0, len(input.Participants))
	for _, p := range input.Participants {
		shares = append(shares, service.ParticipantShare{UserID: p.UserID, Amount: p.Amount})
	}

	bill, err := h.billService.CreateCustomBill(creatorID, input.Title, input.TotalAmount, shares)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Bill custom berhasil dibuat", "bill": bill})
}

// GetBillDetail tetap PUBLIK (tidak protected): melihat detail tagihan
// bersifat seperti membuka link invoice, tidak mengubah data apa pun,
// jadi risikonya rendah dan lebih nyaman dibagikan.
func (h *BillHandler) GetBillDetail(c *gin.Context) {
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

	c.JSON(http.StatusOK, gin.H{"bill": bill, "participants": participants})
}

// SettleParticipant menangani POST /api/bills/participants/:participant_id/settle (PROTECTED).
// user_id yang membayar diambil dari token - mencegah orang melunasi
// (atau mengklaim melunasi) tagihan orang lain.
func (h *BillHandler) SettleParticipant(c *gin.Context) {
	payingUserID, ok := getUserID(c)
	if !ok {
		return
	}

	id := c.Param("participant_id")
	var participantID uint
	if _, err := fmt.Sscanf(id, "%d", &participantID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID peserta tidak valid"})
		return
	}

	if err := h.billService.SettleParticipant(participantID, payingUserID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Pembayaran berhasil, tagihan lunas"})
}