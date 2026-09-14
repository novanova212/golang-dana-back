package service

import (
	"errors"

	"dana-clone/internal/model"
	"dana-clone/internal/repository"
)

type MoneyRequestService interface {
	CreateRequest(requesterID, targetID uint, amount int64, description string) (*model.MoneyRequest, error)
	PayRequest(requestID uint, payingUserID uint) error
	DeclineRequest(requestID uint, decliningUserID uint) error
	GetIncoming(userID uint) ([]model.MoneyRequest, error)
	GetOutgoing(userID uint) ([]model.MoneyRequest, error)
}

type moneyRequestService struct {
	reqRepo       repository.MoneyRequestRepository
	walletService WalletService
}

func NewMoneyRequestService(reqRepo repository.MoneyRequestRepository, walletService WalletService) MoneyRequestService {
	return &moneyRequestService{reqRepo: reqRepo, walletService: walletService}
}

func (s *moneyRequestService) CreateRequest(requesterID, targetID uint, amount int64, description string) (*model.MoneyRequest, error) {
	if amount <= 0 {
		return nil, errors.New("jumlah yang diminta harus lebih dari 0")
	}
	if requesterID == targetID {
		return nil, errors.New("tidak bisa menagih diri sendiri")
	}

	req := &model.MoneyRequest{
		RequesterID: requesterID,
		TargetID:    targetID,
		Amount:      amount,
		Description: description,
		Status:      "pending",
	}
	if err := s.reqRepo.Create(req); err != nil {
		return nil, err
	}
	return req, nil
}

// PayRequest dipanggil ketika TARGET (yang ditagih) setuju membayar.
// Ini memicu Transfer BENERAN dari target ke requester.
func (s *moneyRequestService) PayRequest(requestID uint, payingUserID uint) error {
	req, err := s.reqRepo.FindByID(requestID)
	if err != nil {
		return errors.New("permintaan tidak ditemukan")
	}

	if req.TargetID != payingUserID {
		return errors.New("kamu tidak berhak membayar permintaan ini")
	}
	if req.Status != "pending" {
		return errors.New("permintaan ini sudah tidak berstatus pending")
	}

	note := "Bayar tagihan dari " + req.Requester.Name
	if req.Description != "" {
		note = req.Description
	}

	if err := s.walletService.TransferWithNote(
		payingUserID, req.RequesterID, req.Amount,
		"Bayar: "+note,
		"Diterima dari "+req.Target.Name+": "+note,
	); err != nil {
		return err
	}

	return s.reqRepo.UpdateStatus(requestID, "paid")
}

// DeclineRequest dipanggil ketika target MENOLAK membayar.
func (s *moneyRequestService) DeclineRequest(requestID uint, decliningUserID uint) error {
	req, err := s.reqRepo.FindByID(requestID)
	if err != nil {
		return errors.New("permintaan tidak ditemukan")
	}
	if req.TargetID != decliningUserID {
		return errors.New("kamu tidak berhak menolak permintaan ini")
	}
	if req.Status != "pending" {
		return errors.New("permintaan ini sudah tidak berstatus pending")
	}
	return s.reqRepo.UpdateStatus(requestID, "declined")
}

func (s *moneyRequestService) GetIncoming(userID uint) ([]model.MoneyRequest, error) {
	return s.reqRepo.FindIncoming(userID)
}

func (s *moneyRequestService) GetOutgoing(userID uint) ([]model.MoneyRequest, error) {
	return s.reqRepo.FindOutgoing(userID)
}