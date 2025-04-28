package auditlog

import "github.com/google/uuid"

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
