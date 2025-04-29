package interestrate

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"net/http"
	"payter-bank/features/auditlog"
	"payter-bank/internal/config"
	"payter-bank/internal/database/models"
	databasemocks "payter-bank/internal/database/models/mocks"
	platformerrors "payter-bank/internal/errors"
	"testing"
	"time"
)

func TestService_CreateInterestRate(t *testing.T) {
	t.Run("successfully creates new interest rate when none exists", func(t *testing.T) {
		mocker := newInterestRateMocker(t)

		rateID := uuid.New()
		userID := uuid.New()

		param := CreateInterestRateParam{
			UserID:               userID,
			Rate:                 5.5,
			CalculationFrequency: "monthly",
		}

		// Expected calls
		mocker.db.EXPECT().
			GetInterestRates(gomock.Any()).
			Return(nil, nil)

		mocker.db.EXPECT().
			SaveInterestRate(gomock.Any(), models.SaveInterestRateParams{
				Rate:                 550, // 5.5 * 100
				CalculationFrequency: "monthly",
			}).
			Return(models.InterestRate{
				ID:                   rateID,
				Rate:                 550,
				CalculationFrequency: "monthly",
			}, nil)

		expectedAuditEvent := auditlog.NewEvent(
			auditlog.ActionInterestRateChange,
			userID,
			uuid.Nil,
			auditlog.InterestRateChangeMetadata{
				NewRate:                 550,
				NewCalculationFrequency: "monthly",
			},
		)

		mocker.auditLog.EXPECT().
			Submit(gomock.Any(), expectedAuditEvent).
			Return(nil)

		response, err := mocker.service.CreateInterestRate(context.TODO(), param)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, rateID, response.InterestRateID)
	})

	t.Run("returns existing rate when one exists", func(t *testing.T) {
		mocker := newInterestRateMocker(t)
		existingRateID := uuid.New()

		mocker.db.EXPECT().
			GetInterestRates(gomock.Any()).
			Return([]models.InterestRate{{
				ID:                   existingRateID,
				Rate:                 550,
				CalculationFrequency: "monthly",
			}}, nil)

		response, err := mocker.service.CreateInterestRate(context.Background(), CreateInterestRateParam{
			UserID:               uuid.New(),
			Rate:                 5.5,
			CalculationFrequency: "monthly",
		})

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, existingRateID, response.InterestRateID)
	})
}

func TestService_ApplyRates(t *testing.T) {
	t.Run("successfully applies rates to active accounts with positive balance", func(t *testing.T) {
		mocker := newInterestRateMocker(t)

		rateID := uuid.New()
		account1ID := uuid.New()
		account2ID := uuid.New()
		txnID := uuid.New()

		mocker.db.EXPECT().
			GetInterestRates(gomock.Any()).
			Return([]models.InterestRate{{
				ID:                   rateID,
				Rate:                 500, // 5%
				CalculationFrequency: "monthly",
			}}, nil)

		mocker.db.EXPECT().
			GetAllActiveAccounts(gomock.Any()).
			Return([]models.GetAllActiveAccountsRow{
				{AccountID: account1ID, Currency: "EUR"},
				{AccountID: account2ID, Currency: "USD"},
			}, nil)

		mocker.db.EXPECT().
			GetAccountBalance(gomock.Any(), account1ID).
			Return(models.GetAccountBalanceRow{Balance: 10000}, nil) // 100.00

		mocker.db.EXPECT().
			GetAccountBalance(gomock.Any(), account2ID).
			Return(models.GetAccountBalanceRow{Balance: 20000}, nil) // 200.00

		mocker.db.EXPECT().
			SaveTransaction(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, params models.SaveTransactionParams) (models.Transaction, error) {
				assert.Equal(t, account1ID, params.ToAccountID)
				assert.Equal(t, int64(500), params.Amount) // 5.00 (5% of 100.00)
				assert.Equal(t, "EUR", params.Currency)
				return models.Transaction{ID: txnID}, nil
			}).Times(1)

		mocker.db.EXPECT().
			SaveTransaction(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, params models.SaveTransactionParams) (models.Transaction, error) {
				assert.Equal(t, account2ID, params.ToAccountID)
				assert.Equal(t, int64(1000), params.Amount) // 10.00 (5% of 200.00)
				assert.Equal(t, "USD", params.Currency)
				return models.Transaction{ID: txnID}, nil
			}).Times(1)

		err := mocker.service.ApplyRates(context.Background())

		assert.NoError(t, err)
	})

	t.Run("skips accounts with zero or negative balance", func(t *testing.T) {
		mocker := newInterestRateMocker(t)

		rateID := uuid.New()
		account1ID := uuid.New()
		account2ID := uuid.New()

		mocker.db.EXPECT().
			GetInterestRates(gomock.Any()).
			Return([]models.InterestRate{{
				ID:                   rateID,
				Rate:                 500,
				CalculationFrequency: "monthly",
			}}, nil)

		mocker.db.EXPECT().
			GetAllActiveAccounts(gomock.Any()).
			Return([]models.GetAllActiveAccountsRow{
				{AccountID: account1ID, Currency: "EUR"},
				{AccountID: account2ID, Currency: "USD"},
			}, nil)

		mocker.db.EXPECT().
			GetAccountBalance(gomock.Any(), account1ID).
			Return(models.GetAccountBalanceRow{Balance: 0}, nil)

		mocker.db.EXPECT().
			GetAccountBalance(gomock.Any(), account2ID).
			Return(models.GetAccountBalanceRow{Balance: -1000}, nil)

		err := mocker.service.ApplyRates(context.Background())

		assert.NoError(t, err)
	})

	t.Run("handles errors gracefully", func(t *testing.T) {
		testCases := []struct {
			name          string
			setupMocks    func(*interestRateMocker)
			expectedError error
		}{
			{
				name: "current rate not found",
				setupMocks: func(m *interestRateMocker) {
					m.db.EXPECT().
						GetInterestRates(gomock.Any()).
						Return(nil, nil)
				},
				expectedError: platformerrors.MakeApiError(http.StatusPreconditionFailed, "interest rate has not been initialized"),
			},
			{
				name: "get active accounts error",
				setupMocks: func(m *interestRateMocker) {
					m.db.EXPECT().
						GetInterestRates(gomock.Any()).
						Return([]models.InterestRate{{ID: uuid.New()}}, nil)

					m.db.EXPECT().
						GetAllActiveAccounts(gomock.Any()).
						Return(nil, platformerrors.ErrInternal)
				},
				expectedError: platformerrors.ErrInternal,
			},
			{
				name: "get balance error",
				setupMocks: func(m *interestRateMocker) {
					accountID := uuid.New()

					m.db.EXPECT().
						GetInterestRates(gomock.Any()).
						Return([]models.InterestRate{{ID: uuid.New()}}, nil)

					m.db.EXPECT().
						GetAllActiveAccounts(gomock.Any()).
						Return([]models.GetAllActiveAccountsRow{{AccountID: accountID}}, nil)

					m.db.EXPECT().
						GetAccountBalance(gomock.Any(), accountID).
						Return(models.GetAccountBalanceRow{}, sql.ErrConnDone)
				},
				expectedError: nil, // Should continue with next account
			},
			{
				name: "save transaction error",
				setupMocks: func(m *interestRateMocker) {
					accountID := uuid.New()

					m.db.EXPECT().
						GetInterestRates(gomock.Any()).
						Return([]models.InterestRate{{ID: uuid.New(), Rate: 500}}, nil)

					m.db.EXPECT().
						GetAllActiveAccounts(gomock.Any()).
						Return([]models.GetAllActiveAccountsRow{{AccountID: accountID}}, nil)

					m.db.EXPECT().
						GetAccountBalance(gomock.Any(), accountID).
						Return(models.GetAccountBalanceRow{Balance: 10000}, nil)

					m.db.EXPECT().
						SaveTransaction(gomock.Any(), gomock.Any()).
						Return(models.Transaction{}, sql.ErrTxDone)
				},
				expectedError: nil, // Should continue with next account
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				mocker := newInterestRateMocker(t)
				tc.setupMocks(mocker)

				err := mocker.service.ApplyRates(context.Background())

				if tc.expectedError != nil {
					assert.Equal(t, tc.expectedError, err)
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})
}

func TestService_Start(t *testing.T) {
	t.Run("successfully starts scheduler with correct cron expression", func(t *testing.T) {
		mocker := newInterestRateMocker(t)

		// Setup context with cancellation to test shutdown
		ctx, cancel := context.WithCancel(context.Background())

		mocker.db.EXPECT().
			GetInterestRates(gomock.Any()).
			Return([]models.InterestRate{{
				ID:                   uuid.New(),
				Rate:                 500,
				CalculationFrequency: "daily",
			}}, nil)

		errChan := make(chan error)
		go func() {
			errChan <- mocker.service.Start(ctx)
		}()

		// Allow some time for scheduler to start
		time.Sleep(100 * time.Millisecond)

		// Trigger shutdown
		cancel()

		err := <-errChan
		assert.NoError(t, err)
	})

	t.Run("handles all calculation frequencies", func(t *testing.T) {
		frequencies := map[string]string{
			"hourly":  "0 * * * *",
			"daily":   "0 0 * * *",
			"weekly":  "0 0 * * 0",
			"monthly": "0 0 1 * *",
			"yearly":  "0 0 1 1 *",
		}

		for freq, _ := range frequencies {
			t.Run(fmt.Sprintf("frequency: %s", freq), func(t *testing.T) {
				mocker := newInterestRateMocker(t)
				ctx, cancel := context.WithCancel(context.Background())

				mocker.db.EXPECT().
					GetInterestRates(gomock.Any()).
					Return([]models.InterestRate{{
						ID:                   uuid.New(),
						Rate:                 500,
						CalculationFrequency: freq,
					}}, nil)

				errChan := make(chan error)
				go func() {
					errChan <- mocker.service.Start(ctx)
				}()

				time.Sleep(100 * time.Millisecond)
				cancel()

				err := <-errChan
				assert.NoError(t, err)
			})
		}
	})

	t.Run("handles errors", func(t *testing.T) {
		testCases := []struct {
			name          string
			setupMocks    func(*interestRateMocker)
			expectedError error
		}{
			{
				name: "no current rate",
				setupMocks: func(m *interestRateMocker) {
					m.db.EXPECT().
						GetInterestRates(gomock.Any()).
						Return(nil, nil)
				},
				expectedError: platformerrors.MakeApiError(http.StatusPreconditionFailed, "interest rate has not been initialized"),
			},
			{
				name: "invalid calculation frequency",
				setupMocks: func(m *interestRateMocker) {
					m.db.EXPECT().
						GetInterestRates(gomock.Any()).
						Return([]models.InterestRate{{
							ID:                   uuid.New(),
							Rate:                 500,
							CalculationFrequency: "invalid",
						}}, nil)
				},
				expectedError: fmt.Errorf("unknown frequency: invalid"),
			},
			{
				name: "database error",
				setupMocks: func(m *interestRateMocker) {
					m.db.EXPECT().
						GetInterestRates(gomock.Any()).
						Return(nil, sql.ErrConnDone)
				},
				expectedError: sql.ErrConnDone,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				mocker := newInterestRateMocker(t)
				tc.setupMocks(mocker)

				ctx := context.Background()
				err := mocker.service.Start(ctx)

				assert.Error(t, err)
				assert.Equal(t, tc.expectedError.Error(), err.Error())
			})
		}
	})

	t.Run("graceful shutdown", func(t *testing.T) {
		mocker := newInterestRateMocker(t)
		ctx, cancel := context.WithCancel(context.Background())

		mocker.db.EXPECT().
			GetInterestRates(gomock.Any()).
			Return([]models.InterestRate{{
				ID:                   uuid.New(),
				Rate:                 500,
				CalculationFrequency: "daily",
			}}, nil)

		errChan := make(chan error)
		go func() {
			errChan <- mocker.service.Start(ctx)
		}()

		// Allow scheduler to start
		time.Sleep(100 * time.Millisecond)

		// Test graceful shutdown
		cancel()
		err := <-errChan
		assert.NoError(t, err)
	})

	t.Run("scheduler runs ApplyRates at correct intervals", func(t *testing.T) {
		mocker := newInterestRateMocker(t)
		ctx, cancel := context.WithCancel(context.Background())

		rateID := uuid.New()
		mocker.db.EXPECT().
			GetInterestRates(gomock.Any()).
			Return([]models.InterestRate{{
				ID:                   rateID,
				Rate:                 500,
				CalculationFrequency: "hourly",
			}}, nil).
			AnyTimes()

		mocker.db.EXPECT().
			GetAllActiveAccounts(gomock.Any()).
			Return(nil, nil).
			AnyTimes()

		// Start service
		errChan := make(chan error)
		go func() {
			errChan <- mocker.service.Start(ctx)
		}()

		// Allow some time for at least one execution
		time.Sleep(2 * time.Second)
		cancel()

		err := <-errChan
		assert.NoError(t, err)
	})
}

func TestService_UpdateRate(t *testing.T) {
	t.Run("successfully updates interest rate", func(t *testing.T) {
		mocker := newInterestRateMocker(t)

		rateID := uuid.New()
		userID := uuid.New()

		mocker.db.EXPECT().
			GetInterestRates(gomock.Any()).
			Return([]models.InterestRate{{
				ID:   rateID,
				Rate: 500, // 5%
			}}, nil)

		mocker.db.EXPECT().
			UpdateRate(gomock.Any(), models.UpdateRateParams{
				ID:   rateID,
				Rate: 650, // 6.5%
			}).
			Return(nil)

		expectedAuditEvent := auditlog.NewEvent(
			auditlog.ActionInterestRateChange,
			userID,
			uuid.Nil,
			auditlog.InterestRateChangeMetadata{
				OldRate: 500,
				NewRate: 650,
			},
		)
		mocker.auditLog.EXPECT().
			Submit(gomock.Any(), expectedAuditEvent).
			Return(nil)

		response, err := mocker.service.UpdateRate(context.Background(), UpdateRateParam{
			UserID: userID,
			Rate:   6.5,
		})

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, rateID, response.InterestRateID)
	})

	t.Run("handles non-existent rate", func(t *testing.T) {
		mocker := newInterestRateMocker(t)

		mocker.db.EXPECT().
			GetInterestRates(gomock.Any()).
			Return(nil, nil)

		response, err := mocker.service.UpdateRate(context.Background(), UpdateRateParam{
			UserID: uuid.New(),
			Rate:   6.5,
		})

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, "interest rate has not been initialized", err.Error())
	})

	t.Run("handles database errors", func(t *testing.T) {
		testCases := []struct {
			name          string
			setupMocks    func(*interestRateMocker)
			expectedError error
		}{
			{
				name: "get rates error",
				setupMocks: func(m *interestRateMocker) {
					m.db.EXPECT().
						GetInterestRates(gomock.Any()).
						Return(nil, sql.ErrConnDone)
				},
				expectedError: sql.ErrConnDone,
			},
			{
				name: "update rate error",
				setupMocks: func(m *interestRateMocker) {
					rateID := uuid.New()
					m.db.EXPECT().
						GetInterestRates(gomock.Any()).
						Return([]models.InterestRate{{
							ID:   rateID,
							Rate: 500,
						}}, nil)

					m.db.EXPECT().
						UpdateRate(gomock.Any(), gomock.Any()).
						Return(platformerrors.ErrInternal)
				},
				expectedError: platformerrors.ErrInternal,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				mocker := newInterestRateMocker(t)
				tc.setupMocks(mocker)

				response, err := mocker.service.UpdateRate(context.Background(), UpdateRateParam{
					UserID: uuid.New(),
					Rate:   6.5,
				})

				assert.Error(t, err)
				assert.Nil(t, response)
				assert.Equal(t, tc.expectedError, err)
			})
		}
	})
}

func TestService_UpdateCalculationFrequency(t *testing.T) {
	t.Run("successfully updates calculation frequency", func(t *testing.T) {
		mocker := newInterestRateMocker(t)

		rateID := uuid.New()
		userID := uuid.New()

		mocker.db.EXPECT().
			GetInterestRates(gomock.Any()).
			Return([]models.InterestRate{{
				ID:                   rateID,
				Rate:                 500,
				CalculationFrequency: "monthly",
			}}, nil)

		mocker.db.EXPECT().
			UpdateCalculationFrequency(gomock.Any(), models.UpdateCalculationFrequencyParams{
				ID:                   rateID,
				CalculationFrequency: "weekly",
			}).
			Return(nil)

		expectedAuditEvent := auditlog.NewEvent(
			auditlog.ActionInterestRateChange,
			userID,
			uuid.Nil,
			auditlog.InterestRateChangeMetadata{
				OldCalculationFrequency: "monthly",
				NewCalculationFrequency: "weekly",
			},
		)
		mocker.auditLog.EXPECT().
			Submit(gomock.Any(), expectedAuditEvent).
			Return(nil)

		response, err := mocker.service.UpdateCalculationFrequency(context.Background(), UpdateCalculationFrequencyParam{
			UserID:               userID,
			CalculationFrequency: "weekly",
		})

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, rateID, response.InterestRateID)
	})

	t.Run("handles non-existent rate", func(t *testing.T) {
		mocker := newInterestRateMocker(t)

		mocker.db.EXPECT().
			GetInterestRates(gomock.Any()).
			Return(nil, nil)

		response, err := mocker.service.UpdateCalculationFrequency(context.Background(), UpdateCalculationFrequencyParam{
			UserID:               uuid.New(),
			CalculationFrequency: "monthly",
		})

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, "interest rate has not been initialized", err.Error())
	})

	t.Run("handles database errors", func(t *testing.T) {
		testCases := []struct {
			name          string
			setupMocks    func(*interestRateMocker)
			expectedError error
		}{
			{
				name: "get rates error",
				setupMocks: func(m *interestRateMocker) {
					m.db.EXPECT().
						GetInterestRates(gomock.Any()).
						Return(nil, sql.ErrConnDone)
				},
				expectedError: sql.ErrConnDone,
			},
			{
				name: "update frequency error",
				setupMocks: func(m *interestRateMocker) {
					rateID := uuid.New()
					m.db.EXPECT().
						GetInterestRates(gomock.Any()).
						Return([]models.InterestRate{{
							ID:                   rateID,
							CalculationFrequency: "monthly",
						}}, nil)

					m.db.EXPECT().
						UpdateCalculationFrequency(gomock.Any(), gomock.Any()).
						Return(platformerrors.ErrInternal)
				},
				expectedError: platformerrors.ErrInternal,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				mocker := newInterestRateMocker(t)
				tc.setupMocks(mocker)

				response, err := mocker.service.UpdateCalculationFrequency(context.Background(), UpdateCalculationFrequencyParam{
					UserID:               uuid.New(),
					CalculationFrequency: "weekly",
				})

				assert.Error(t, err)
				assert.Nil(t, response)
				assert.Equal(t, tc.expectedError, err)
			})
		}
	})

	t.Run("prevents update to same frequency", func(t *testing.T) {
		mocker := newInterestRateMocker(t)
		rateID := uuid.New()
		rate := models.InterestRate{
			ID:                   rateID,
			CalculationFrequency: "monthly",
		}

		mocker.db.EXPECT().
			GetInterestRates(gomock.Any()).
			Return([]models.InterestRate{rate}, nil)

		response, err := mocker.service.UpdateCalculationFrequency(context.Background(), UpdateCalculationFrequencyParam{
			UserID:               uuid.New(),
			CalculationFrequency: "monthly",
		})

		assert.NoError(t, err)
		assert.Equal(t, &Response{InterestRateID: rateID}, response)
	})
}

type interestRateMocker struct {
	db       *databasemocks.MockQuerier
	auditLog *auditlog.MockService

	service Service
}

func newInterestRateMocker(t *testing.T) *interestRateMocker {
	ctrl := gomock.NewController(t)
	db := databasemocks.NewMockQuerier(ctrl)
	auditLog := auditlog.NewMockService(ctrl)
	cfg := config.AppConfig{
		InterestRateAccountID: uuid.MustParse("00000000-0000-0000-0000-000000000000"),
	}

	svc := NewService(db, cfg, auditLog)
	return &interestRateMocker{
		db:       db,
		auditLog: auditLog,
		service:  svc,
	}
}
