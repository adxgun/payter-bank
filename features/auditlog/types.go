package auditlog

import (
	"github.com/google/uuid"
	"payter-bank/internal/database/models"
	"time"
)

type Event struct {
	Action    Action    `json:"action"`
	UserID    uuid.UUID `json:"user_id"`
	AccountID uuid.UUID `json:"account_id"`
	Metadata  any       `json:"metadata"`
}

func NewEvent(action Action, userID, accountID uuid.UUID, metadata any) Event {
	return Event{
		Action:    action,
		UserID:    userID,
		AccountID: accountID,
		Metadata:  metadata,
	}
}

var auditLogTaskName = "auditlog:record"

type Action string

const (
	ActionCreateAccount       Action = "create_account"
	ActionAccountStatusChange Action = "account_status_change"
	ActionAccountCredit       Action = "account_credit"
	ActionAccountDebit        Action = "account_debit"
	ActionAccountTransfer     Action = "account_transfer"
	ActionInterestRateChange  Action = "interest_rate_change"
)

func (a Action) String() string {
	return string(a)
}

type AccountStatusChangeMetadata struct {
	OldStatus string `json:"old_status"`
	NewStatus string `json:"new_status"`
}

type InterestRateChangeMetadata struct {
	OldRate                 int64  `json:"old_rate"`
	OldCalculationFrequency string `json:"old_calculation_frequency"`
	NewRate                 int64  `json:"new_rate"`
	NewCalculationFrequency string `json:"new_calculation_frequency"`
}

type AuditLog struct {
	AccountID     uuid.UUID   `json:"account_id"`
	CurrentStatus string      `json:"current_status"`
	ActionCode    string      `json:"action_code"`
	Action        interface{} `json:"action"`
	OldStatus     string      `json:"old_status"`
	NewStatus     string      `json:"new_status"`
	Amount        Amount      `json:"amount"`
	ActionBy      interface{} `json:"action_by"`
	CreatedAt     time.Time   `json:"created_at"`
}

type Amount struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

func AuditLogFromRow(row models.GetAuditLogsForAccountRow) AuditLog {
	return AuditLog{
		AccountID:     row.AccountID,
		CurrentStatus: string(row.CurrentStatus),
		ActionCode:    row.ActionCode,
		Action:        row.Action,
		OldStatus:     row.OldStatus,
		NewStatus:     row.NewStatus,
		Amount: Amount{
			Amount:   float64(row.Amount / 100),
			Currency: row.Currency,
		},
		ActionBy:  row.ActionBy,
		CreatedAt: row.CreatedAt.Time,
	}
}
