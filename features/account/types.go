package account

import (
	"github.com/google/uuid"
	"payter-bank/internal/database/models"
	"time"
)

type Profile struct {
	AccountID    uuid.UUID `json:"account_id"`
	UserID       uuid.UUID `json:"user_id"`
	Email        string    `json:"email"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	AccountType  string    `json:"account_type"`
	UserType     string    `json:"user_type"`
	RegisteredAt time.Time `json:"registered_at"`
}

func ProfileFromQueryResult(r models.GetProfileByUserIDRow) Profile {
	return Profile{
		AccountID:    r.AccountID,
		UserID:       r.UserID,
		Email:        r.Email,
		FirstName:    r.FirstName,
		LastName:     r.LastName,
		AccountType:  string(r.AccountType),
		RegisteredAt: r.RegisteredAt.Time,
		UserType:     string(r.UserType),
	}
}

type AuthenticateAccountParams struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type CreateUserParams struct {
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required"`
	UserType  string `json:"user_type" binding:"required,oneof=CUSTOMER ADMIN"`
}

type CreateUserResponse struct {
	UserID uuid.UUID `json:"user_id"`
}

type CreateAccountParams struct {
	Currency       string    `json:"currency" binding:"required,oneof=GBP EUR JPY"`
	InitialDeposit float64   `json:"initial_deposit"`
	UserID         uuid.UUID `json:"user_id" binding:"required"`
	AdminUserID    uuid.UUID
}

type OperationParams struct {
	UserID    uuid.UUID
	AccountID uuid.UUID
}

type AccessToken struct {
	Token string `json:"token"`
}

type ChangeHistory struct {
	AccountID     uuid.UUID `json:"account_id"`
	CurrentStatus string    `json:"current_status"`
	Action        string    `json:"action"`
	OldStatus     string    `json:"old_status"`
	NewStatus     string    `json:"new_status"`
	ActionBy      string    `json:"action_by"`
	CreatedAt     time.Time `json:"created_at"`
}

func ChangeHistoryFromRow(row models.GetAccountStatusHistoryRow) ChangeHistory {
	return ChangeHistory{
		AccountID:     row.AccountID,
		CurrentStatus: string(row.CurrentStatus),
		Action:        row.Action,
		OldStatus:     row.OldStatus,
		NewStatus:     row.NewStatus,
		ActionBy:      row.ActionBy.(string),
		CreatedAt:     row.CreatedAt.Time,
	}
}

type Amount struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

type Account struct {
	UserID        uuid.UUID `json:"user_id"`
	AccountID     uuid.UUID `json:"account_id"`
	AccountNumber string    `json:"account_number"`
	AccountType   string    `json:"account_type"`
	Currency      string    `json:"currency"`
	Balance       Amount    `json:"balance"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	FirstName     string    `json:"first_name"`
	LastName      string    `json:"last_name"`
}

func AccountFromQuery(row models.GetAllCurrentAccountsRow) Account {
	return Account{
		UserID:        row.UserID,
		AccountID:     row.AccountID,
		AccountNumber: row.AccountNumber,
		AccountType:   string(row.AccountType),
		Currency:      string(row.Status),
		Balance: Amount{
			Amount:   float64(row.Balance.Int64) / 100,
			Currency: string(row.Currency),
		},
		Status:    string(row.Status),
		CreatedAt: row.CreatedAt.Time,
		FirstName: row.FirstName,
		LastName:  row.LastName,
	}
}

func AccountFromDetailsRow(row models.GetAccountDetailsByIDRow) Account {
	return Account{
		UserID:        row.UserID,
		AccountID:     row.AccountID,
		AccountNumber: row.AccountNumber,
		AccountType:   string(row.AccountType),
		Currency:      string(row.Currency),
		Balance: Amount{
			Amount:   float64(row.Balance.Int64) / 100,
			Currency: string(row.Currency),
		},
		FirstName: row.FirstName,
		LastName:  row.LastName,
		Status:    string(row.Status),
		CreatedAt: row.CreatedAt.Time,
	}
}
