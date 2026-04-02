// internal/dto/dto.go  (updated with validate tags)
package dto

type CreateTransactionInput struct {
	UserID      string  `json:"user_id"`
	Amount      float64 `json:"amount"      validate:"required,gt=0"`
	Type        string  `json:"type"        validate:"required,oneof=income expense"`
	Category    string  `json:"category"    validate:"required,min=1,max=50"`
	Description string  `json:"description" validate:"max=255"`
	Date        string  `json:"date"        validate:"required"`
}

type UpdateTransactionInput struct {
	Amount      float64 `json:"amount"      validate:"required,gt=0"`
	Type        string  `json:"type"        validate:"required,oneof=income expense"`
	Category    string  `json:"category"    validate:"required,min=1,max=50"`
	Description string  `json:"description" validate:"max=255"`
	Date        string  `json:"date"        validate:"required"`
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

type RegisterInput struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Name     string `json:"name"     validate:"required,min=2"`
	Role     string `json:"role"     validate:"omitempty,oneof=viewer analyst admin"`
}

type LoginInput struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}
