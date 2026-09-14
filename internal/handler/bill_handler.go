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

// CreateBill menangani POST /api/bills (bagi RATA, cara lama).
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

	c.JSON(http.StatusCreated, gin.H{"message": "Bill berhasil dibuat", "bill": bill})
}

// ParticipantShareInput mewakili satu baris input custom split dari client.
type ParticipantShareInput struct {
	UserID uint  `json:"user_id" binding:"required"`
	Amount int64 `json:"amount" binding:"required"`
}

type CreateCustomBillInput struct {
	CreatorID    uint                    `json:"creator_id" binding:"required"`
	Title        string                  `json:"title" binding:"required"`
	TotalAmount  int64                   `json:"total_amount" binding:"required"`
	Participants []ParticipantShareInput `json:"participants" binding:"required"`
}

// CreateCustomBill menangani POST /api/bills/custom (porsi BEDA-BEDA per orang).
func (h *BillHandler) CreateCustomBill(c *gin.Context) {
	var input CreateCustomBillInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	shares := make([]service.ParticipantShare, 0, len(input.Participants))
	for _, p := range input.Participants {
		shares = append(shares, service.ParticipantShare{UserID: p.UserID, Amount: p.Amount})
	}

	bill, err := h.billService.CreateCustomBill(input.CreatorID, input.Title, input.TotalAmount, shares)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Bill custom berhasil dibuat", "bill": bill})
}

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

	creatorShare := service.CalculateCreatorShare(bill, participants)

	// Hitung juga progres pelunasan: berapa dari total yang sudah masuk.
	var totalPaid int64 = creatorShare // porsi creator dianggap "sudah dibayar" karena dia yang bayar duluan
	for _, p := range participants {
		if p.Paid {
			totalPaid += p.Amount
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"bill":          bill,
		"participants":  participants,
		"creator_share": creatorShare,
		"total_paid":    totalPaid,
	})
}

type SettleInput struct {
	UserID uint `json:"user_id" binding:"required"`
}

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