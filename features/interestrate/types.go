package interestrate

import "github.com/google/uuid"

type Frequency string

const (
	Hourly  Frequency = "hourly"
	Daily   Frequency = "daily"
	Weekly  Frequency = "weekly"
	Monthly Frequency = "monthly"
	Yearly  Frequency = "yearly"
)

type CreateInterestRateParam struct {
	UserID               uuid.UUID
	Rate                 float64 `json:"rate" binding:"required,gt=0"`
	CalculationFrequency string  `json:"calculation_frequency" binding:"required,oneof=hourly daily weekly monthly yearly"`
}

type UpdateRateParam struct {
	UserID uuid.UUID
	Rate   float64 `json:"rate" binding:"required,gte=0"`
}

type UpdateCalculationFrequencyParam struct {
	UserID               uuid.UUID
	CalculationFrequency string `json:"calculation_frequency" binding:"required,oneof=hourly daily weekly monthly yearly"`
}

type Response struct {
	InterestRateID uuid.UUID `json:"interest_rate_id"`
}
