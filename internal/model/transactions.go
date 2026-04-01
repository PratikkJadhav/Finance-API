package model

import (
	"time"

	"github.com/google/uuid"
)

type TxnType string

const (
	TxnTypeIncome  TxnType = "income"
	TxnTypeExpense TxnType = "expense"
)

type Transaction struct {
	ID          uuid.UUID  `json:"id"`
	UserID      uuid.UUID  `json:"user_id"`
	Amount      float64    `json:"amount"`
	Type        TxnType    `json:"type"`
	Category    string     `json:"category"`
	Description *string    `json:"description,omitempty"` // Pointer to handle SQL NULL
	Date        time.Time  `json:"date"`
	DeletedAt   *time.Time `json:"-"` // Hide soft deletes from frontend
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}
