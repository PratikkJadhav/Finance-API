package service

import (
	"context"
	"errors"

	"github.com/PratikkJadhav/Finance-API/internal/dto"
	"github.com/PratikkJadhav/Finance-API/internal/model"
	"github.com/PratikkJadhav/Finance-API/internal/repository"
	"github.com/PratikkJadhav/Finance-API/internal/validator"
)

type TransactionService struct {
	txnRepo *repository.TransactionRepo
}

func NewTransactionService(txnRepo *repository.TransactionRepo) *TransactionService {
	return &TransactionService{txnRepo: txnRepo}
}

type CreateTransactionInput struct {
	UserID      string  `json:"user_id"`
	Amount      float64 `json:"amount"`
	Type        string  `json:"type"`
	Category    string  `json:"category"`
	Description string  `json:"description"`
	Date        string  `json:"date"` // expect "YYYY-MM-DD"
}

type UpdateTransactionInput struct {
	Amount      float64 `json:"amount"`
	Type        string  `json:"type"`
	Category    string  `json:"category"`
	Description string  `json:"description"`
	Date        string  `json:"date"`
}

type TransactionFilter struct {
	UserID   string
	Type     string
	Category string
	From     string
	To       string
	Page     int
	Limit    int
}

func (s *TransactionService) Create(ctx context.Context, input dto.CreateTransactionInput) (*model.Transaction, error) {
	if err := validator.Validate(input); err != nil {
		return nil, err
	}
	if input.Amount <= 0 {
		return nil, errors.New("amount must be greater than 0")
	}
	if input.Type != "income" && input.Type != "expense" {
		return nil, errors.New("type must be income or expense")
	}
	if input.Category == "" {
		return nil, errors.New("category is required")
	}
	if input.Date == "" {
		return nil, errors.New("date is required")
	}

	return s.txnRepo.Create(ctx, input)
}

func (s *TransactionService) List(ctx context.Context, filter dto.TransactionFilter) ([]*model.Transaction, int, error) {
	return s.txnRepo.GetAll(ctx, filter)
}

func (s *TransactionService) GetByID(ctx context.Context, id string) (*model.Transaction, error) {
	return s.txnRepo.GetByID(ctx, id)
}

func (s *TransactionService) Update(ctx context.Context, id string, input dto.UpdateTransactionInput) (*model.Transaction, error) {
	if input.Amount <= 0 {
		return nil, errors.New("amount must be greater than 0")
	}
	if input.Type != "income" && input.Type != "expense" {
		return nil, errors.New("type must be income or expense")
	}

	return s.txnRepo.Update(ctx, id, input)
}

func (s *TransactionService) Delete(ctx context.Context, id string) error {
	return s.txnRepo.SoftDelete(ctx, id)
}

func (s *TransactionService) GetSummary(ctx context.Context, userID string, role string) (map[string]interface{}, error) {
	// admins see global summary, others see their own
	if role == "admin" {
		return s.txnRepo.GetGlobalSummary(ctx)
	}
	return s.txnRepo.GetSummaryByUser(ctx, userID)
}

func (s *TransactionService) GetTrends(ctx context.Context, userID string, role string) (interface{}, error) {
	if role == "admin" {
		return s.txnRepo.GetMonthlyTrends(ctx, "")
	}
	return s.txnRepo.GetMonthlyTrends(ctx, userID)
}

func (s *TransactionService) GetCategoryTotals(ctx context.Context, userID string, role string) (interface{}, error) {
	if role == "admin" {
		return s.txnRepo.GetCategoryTotals(ctx, "")
	}
	return s.txnRepo.GetCategoryTotals(ctx, userID)
}

func (s *TransactionService) GetRecent(ctx context.Context, userID string, role string) ([]*model.Transaction, error) {
	return s.txnRepo.GetRecent(ctx, userID, role)
}
