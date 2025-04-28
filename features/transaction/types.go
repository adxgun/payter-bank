package transaction

import (
	"github.com/google/uuid"
	"payter-bank/internal/database/models"
	"time"
)

type AccountTransactionParams struct {
	FromAccountID uuid.UUID `json:"from_account_id"`
	ToAccountID   uuid.UUID `json:"to_account_id"`
	Amount        float64   `json:"amount" binding:"required"`
	Narration     string    `json:"narration"`
	UserID        uuid.UUID
}

func (p AccountTransactionParams) AmountUnit() int64 {
	return int64(p.Amount * 100)
}

type Response struct {
	TransactionID uuid.UUID `json:"transaction_id"`
}

type Balance struct {
	AccountID     uuid.UUID `json:"account_id"`
	Balance       float64   `json:"balance"`
	AccountNumber string    `json:"account_number"`
	AccountType   string    `json:"account_type"`
	Currency      string    `json:"currency"`
}

func BalanceFromQueryResult(balance models.GetAccountBalanceRow) Balance {
	return Balance{
		AccountID:     balance.AccountID,
		Balance:       float64(balance.Balance) / 100,
		AccountNumber: balance.AccountNumber,
		AccountType:   string(balance.AccountType),
		Currency:      string(balance.Currency),
	}
}

type Amount struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

type Transaction struct {
	TransactionID   uuid.UUID `json:"transaction_id"`
	FromAccountID   uuid.UUID `json:"from_account_id"`
	ToAccountID     uuid.UUID `json:"to_account_id"`
	Amount          Amount    `json:"amount"`
	ReferenceNumber string    `json:"reference_number"`
	Description     string    `json:"description"`
	Status          string    `json:"status"`
	Currency        string    `json:"currency"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func TransactionFromRow(r models.GetTransactionsByAccountIDRow) Transaction {
	return Transaction{
		TransactionID: r.TransactionID,
		FromAccountID: r.FromAccountID,
		ToAccountID:   r.ToAccountID,
		Amount: Amount{
			Amount:   float64(r.Amount) / 100,
			Currency: r.Currency,
		},
		ReferenceNumber: r.ReferenceNumber,
		Description:     r.Description.String,
		Status:          r.Status,
		Currency:        r.Currency,
		CreatedAt:       r.CreatedAt.Time,
		UpdatedAt:       r.UpdatedAt.Time,
	}
}
