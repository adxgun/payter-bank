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

type CreateAccountParams struct {
	FirstName      string  `json:"first_name" binding:"required"`
	LastName       string  `json:"last_name" binding:"required"`
	Email          string  `json:"email" binding:"required,email"`
	Password       string  `json:"password" binding:"required"`
	Currency       string  `json:"currency" binding:"required,oneof=GBP EUR JPY"`
	UserType       string  `json:"user_type" binding:"required,oneof=CUSTOMER ADMIN"`
	InitialDeposit float64 `json:"initial_deposit"`
	UserID         uuid.UUID
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
