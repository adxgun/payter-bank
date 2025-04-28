package transaction

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"payter-bank/internal/database/models"
	"testing"
)

func TestBalanceFromQueryResult(t *testing.T) {
	t.Run("successfully converts positive balance", func(t *testing.T) {
		input := models.GetAccountBalanceRow{
			AccountID:     uuid.New(),
			Balance:       15000, // 150.00
			AccountNumber: "1234567890",
			AccountType:   models.AccountTypeCURRENT,
			Currency:      models.CurrencyGBP,
		}

		expected := Balance{
			AccountID:     input.AccountID,
			Balance:       150.00,
			AccountNumber: input.AccountNumber,
			AccountType:   string(input.AccountType),
			Currency:      string(input.Currency),
		}

		result := BalanceFromQueryResult(input)
		assert.Equal(t, expected, result)
	})

	t.Run("successfully converts negative balance", func(t *testing.T) {
		input := models.GetAccountBalanceRow{
			AccountID:     uuid.New(),
			Balance:       -5000, // -50.00
			AccountNumber: "1234567890",
			AccountType:   models.AccountTypeCURRENT,
			Currency:      models.CurrencyGBP,
		}

		expected := Balance{
			AccountID:     input.AccountID,
			Balance:       -50.00,
			AccountNumber: input.AccountNumber,
			AccountType:   string(input.AccountType),
			Currency:      string(input.Currency),
		}

		result := BalanceFromQueryResult(input)
		assert.Equal(t, expected, result)
	})

	t.Run("successfully converts zero balance", func(t *testing.T) {
		input := models.GetAccountBalanceRow{
			AccountID:     uuid.New(),
			Balance:       0,
			AccountNumber: "1234567890",
			AccountType:   models.AccountTypeCURRENT,
			Currency:      models.CurrencyGBP,
		}

		expected := Balance{
			AccountID:     input.AccountID,
			Balance:       0,
			AccountNumber: input.AccountNumber,
			AccountType:   string(input.AccountType),
			Currency:      string(input.Currency),
		}

		result := BalanceFromQueryResult(input)
		assert.Equal(t, expected, result)
	})

	t.Run("handles decimal conversion correctly", func(t *testing.T) {
		testCases := []struct {
			balance     int64
			expected    float64
			description string
		}{
			{100, 1.00, "simple conversion"},
			{1, 0.01, "smallest unit"},
			{99999999, 999999.99, "large number"},
			{-100, -1.00, "negative number"},
			{50, 0.50, "half unit"},
			{10, 0.10, "tenth unit"},
		}

		for _, tc := range testCases {
			t.Run(tc.description, func(t *testing.T) {
				input := models.GetAccountBalanceRow{
					AccountID:     uuid.New(),
					Balance:       int32(tc.balance),
					AccountNumber: "1234567890",
					AccountType:   models.AccountTypeCURRENT,
					Currency:      models.CurrencyGBP,
				}

				result := BalanceFromQueryResult(input)
				assert.Equal(t, tc.expected, result.Balance)
			})
		}
	})
}
