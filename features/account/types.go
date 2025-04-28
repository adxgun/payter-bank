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
	}
}
