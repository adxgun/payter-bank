package account

import (
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"payter-bank/features/auditlog"
	"payter-bank/features/transaction"
	"payter-bank/internal/database/models"
	databasemocks "payter-bank/internal/database/models/mocks"
	platformerrors "payter-bank/internal/errors"
	"payter-bank/internal/pkg/generator"
	generatormocks "payter-bank/internal/pkg/generator/mocks"
	"payter-bank/internal/pkg/password"
	passwordhashermocks "payter-bank/internal/pkg/password/mocks"
	"testing"
)

func TestService_InitialiseAdmin(t *testing.T) {
	t.Run("initialise admin successfully", func(t *testing.T) {
		m := mockAccountService(t)
		email, pwd := "testmail@test.com", "password"
		userID, accountID, accountNumber := uuid.New(), uuid.New(), "00001111"
		expectedSaveUserParams := models.SaveUserParams{
			Email:     email,
			Password:  "hashedPassword",
			FirstName: "Admin",
			LastName:  "Admin",
			UserType:  "ADMIN",
		}
		expectedSaveAccountParams := models.SaveAccountParams{
			UserID:        userID,
			AccountNumber: accountNumber,
			Status:        "ACTIVE",
			AccountType:   "CURRENT",
			Currency:      "GBP",
		}

		m.db.EXPECT().GetUserByEmail(gomock.Any(), email).
			Return(models.GetUserByEmailRow{}, nil)
		m.numberGenerator.EXPECT().Generate().Return("00001111")
		m.passwordHasher.EXPECT().Hash("password").Return("hashedPassword")
		m.db.EXPECT().SaveUser(gomock.Any(), expectedSaveUserParams).
			Return(models.SaveUserRow{ID: userID}, nil)
		m.db.EXPECT().SaveAccount(gomock.Any(), expectedSaveAccountParams).
			Return(models.Account{ID: accountID}, nil)

		err := m.service.InitialiseAdmin(context.TODO(), email, pwd)
		assert.Nil(t, err)
	})

	t.Run("fails when admin user already exists", func(t *testing.T) {
		m := mockAccountService(t)
		email, pwd := "admin@test.com", "password"

		m.db.EXPECT().GetUserByEmail(gomock.Any(), email).
			Return(models.GetUserByEmailRow{
				ID:       uuid.New(),
				Email:    email,
				UserType: "ADMIN",
			}, nil)

		err := m.service.InitialiseAdmin(context.TODO(), email, pwd)
		assert.NoError(t, err)
	})
}

func TestService_AuthenticateAccount(t *testing.T) {
	t.Run("successful authentication", func(t *testing.T) {
		m := mockAccountService(t)
		email, pwd := "test@example.com", "password"
		userID, accountID := uuid.New(), uuid.New()
		generatedToken := uuid.NewString()
		expectedToken := AccessToken{
			Token: generatedToken,
		}
		profile := models.GetProfileByUserIDRow{
			AccountID: accountID,
			UserID:    userID,
		}

		m.db.EXPECT().GetUserByEmail(gomock.Any(), email).
			Return(models.GetUserByEmailRow{
				ID:       userID,
				Email:    email,
				Password: "hashedPassword",
				UserType: "ADMIN",
			}, nil)
		m.passwordHasher.EXPECT().Validate("hashedPassword", "password").
			Return(true)
		m.db.EXPECT().GetProfileByUserID(gomock.Any(), userID).
			Return(profile, nil)

		tokenData := generator.TokenData{
			UserID:    userID,
			AccountID: accountID,
		}
		m.generator.EXPECT().Generate(tokenData).Return(generatedToken, nil)

		result, err := m.service.AuthenticateAccount(context.TODO(), AuthenticateAccountParams{
			Email:    email,
			Password: pwd,
		})

		assert.NoError(t, err)
		assert.Equal(t, expectedToken, result)
	})

	t.Run("fails when user not found", func(t *testing.T) {
		m := mockAccountService(t)
		email, pwd := "nonexistent@example.com", "password"

		m.db.EXPECT().GetUserByEmail(gomock.Any(), email).
			Return(models.GetUserByEmailRow{}, sql.ErrNoRows)

		result, err := m.service.AuthenticateAccount(context.TODO(), AuthenticateAccountParams{
			Email:    email,
			Password: pwd,
		})

		assert.Error(t, err)
		assert.Empty(t, result)
		assert.Contains(t, err.Error(), "invalid login credentials")
	})

	t.Run("fails when password is incorrect", func(t *testing.T) {
		m := mockAccountService(t)
		email, pwd := "test@example.com", "wrongpassword"
		userID := uuid.New()

		m.db.EXPECT().GetUserByEmail(gomock.Any(), email).
			Return(models.GetUserByEmailRow{
				ID:       userID,
				Email:    email,
				Password: "hashedPassword",
				UserType: "ADMIN",
			}, nil)

		m.passwordHasher.EXPECT().Validate("hashedPassword", "wrongpassword").
			Return(false)

		result, err := m.service.AuthenticateAccount(context.TODO(), AuthenticateAccountParams{
			Email:    email,
			Password: pwd,
		})

		assert.Error(t, err)
		assert.Empty(t, result)
		assert.Contains(t, err.Error(), "invalid login credentials")
	})

	t.Run("fails when token generation fails", func(t *testing.T) {
		m := mockAccountService(t)
		email, pwd := "test@example.com", "password"
		userID, accountID := uuid.New(), uuid.New()
		profile := models.GetProfileByUserIDRow{
			AccountID: accountID,
			UserID:    userID,
		}

		m.db.EXPECT().GetUserByEmail(gomock.Any(), email).
			Return(models.GetUserByEmailRow{
				ID:       userID,
				Email:    email,
				Password: "hashedPassword",
				UserType: "ADMIN",
			}, nil)

		m.passwordHasher.EXPECT().Validate("hashedPassword", "password").
			Return(true)

		m.generator.EXPECT().Generate(generator.TokenData{
			AccountID: accountID,
			UserID:    userID,
		}).Return("", errors.New("failed to generate token"))
		m.db.EXPECT().GetProfileByUserID(gomock.Any(), userID).
			Return(profile, nil)

		result, err := m.service.AuthenticateAccount(context.TODO(), AuthenticateAccountParams{
			Email:    email,
			Password: pwd,
		})

		assert.Error(t, err)
		assert.Empty(t, result)
		assert.Contains(t, err.Error(), platformerrors.ErrInternal.Error())
	})
}

func TestService_GetProfile(t *testing.T) {
	t.Run("successfully gets user profile", func(t *testing.T) {
		m := mockAccountService(t)
		userID := uuid.New()
		accountID := uuid.New()
		profile := models.GetProfileByUserIDRow{
			AccountID: accountID,
			UserID:    userID,
		}
		expectedProfile := Profile{
			AccountID: accountID,
			UserID:    userID,
		}

		m.db.EXPECT().GetProfileByUserID(gomock.Any(), userID).
			Return(profile, nil)

		parsedProfile, err := m.service.GetProfile(context.TODO(), userID)

		assert.NoError(t, err)
		assert.Equal(t, expectedProfile, parsedProfile)
	})

	t.Run("fails when user not found", func(t *testing.T) {
		m := mockAccountService(t)
		userID := uuid.New()

		m.db.EXPECT().GetProfileByUserID(gomock.Any(), userID).
			Return(models.GetProfileByUserIDRow{}, sql.ErrNoRows)

		profile, err := m.service.GetProfile(context.TODO(), userID)

		assert.Error(t, err)
		assert.Empty(t, profile)
		assert.Contains(t, err.Error(), "profile not found")
	})

	t.Run("fails when getting accounts fails", func(t *testing.T) {
		m := mockAccountService(t)
		userID := uuid.New()

		m.db.EXPECT().GetProfileByUserID(gomock.Any(), userID).
			Return(models.GetProfileByUserIDRow{}, errors.New("database error"))

		profile, err := m.service.GetProfile(context.TODO(), userID)

		assert.Error(t, err)
		assert.Empty(t, profile)
		assert.Contains(t, err.Error(), platformerrors.ErrInternal.Error())
	})
}

func TestService_SuspendAccount(t *testing.T) {
	t.Run("successfully suspends account", func(t *testing.T) {
		m := mockAccountService(t)
		accountID := uuid.New()
		userID := uuid.New()
		account := models.GetAccountByIDRow{
			ID:     accountID,
			UserID: userID,
			Status: "ACTIVE",
		}
		expectedEventLog := auditlog.Event{
			Action:    auditlog.ActionAccountStatusChange,
			UserID:    userID,
			AccountID: accountID,
			Metadata: auditlog.AccountStatusChangeMetadata{
				OldStatus: "ACTIVE",
				NewStatus: "SUSPENDED",
			},
		}

		m.db.EXPECT().GetAccountByID(gomock.Any(), accountID).
			Return(account, nil)
		m.db.EXPECT().UpdateAccountStatus(gomock.Any(), models.UpdateAccountStatusParams{
			ID:     accountID,
			Status: "SUSPENDED",
		}).Return(nil)

		m.auditLog.EXPECT().Submit(gomock.Any(), expectedEventLog).
			Return(nil)

		err := m.service.SuspendAccount(context.TODO(), OperationParams{
			UserID:    userID,
			AccountID: accountID,
		})
		assert.NoError(t, err)
	})

	t.Run("fails when account not found", func(t *testing.T) {
		m := mockAccountService(t)
		accountID, userID := uuid.New(), uuid.New()

		m.db.EXPECT().GetAccountByID(gomock.Any(), accountID).
			Return(models.GetAccountByIDRow{}, sql.ErrNoRows)

		err := m.service.SuspendAccount(context.TODO(), OperationParams{
			UserID:    userID,
			AccountID: accountID,
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "account not found")
	})

	t.Run("fails when account is already suspended", func(t *testing.T) {
		m := mockAccountService(t)
		accountID := uuid.New()
		userID := uuid.New()
		account := models.GetAccountByIDRow{
			ID:            accountID,
			UserID:        userID,
			AccountNumber: "1234567890",
			Status:        "SUSPENDED",
		}

		m.db.EXPECT().GetAccountByID(gomock.Any(), accountID).
			Return(account, nil)

		err := m.service.SuspendAccount(context.TODO(), OperationParams{
			UserID:    userID,
			AccountID: accountID,
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "account already suspended")
	})

	t.Run("fails when update status fails", func(t *testing.T) {
		m := mockAccountService(t)
		accountID := uuid.New()
		userID := uuid.New()
		account := models.GetAccountByIDRow{
			ID:     accountID,
			UserID: userID,
		}

		m.db.EXPECT().GetAccountByID(gomock.Any(), accountID).
			Return(account, nil)

		m.db.EXPECT().UpdateAccountStatus(gomock.Any(), models.UpdateAccountStatusParams{
			ID:     accountID,
			Status: "SUSPENDED",
		}).Return(errors.New("database error"))

		err := m.service.SuspendAccount(context.TODO(), OperationParams{
			UserID:    userID,
			AccountID: accountID,
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), platformerrors.ErrInternal.Error())
	})
}

func TestService_ActivateAccount(t *testing.T) {
	t.Run("successfully activate account", func(t *testing.T) {
		m := mockAccountService(t)
		accountID := uuid.New()
		userID := uuid.New()
		account := models.GetAccountByIDRow{
			ID:     accountID,
			UserID: userID,
			Status: "SUSPENDED",
		}
		expectedEventLog := auditlog.Event{
			Action:    auditlog.ActionAccountStatusChange,
			UserID:    userID,
			AccountID: accountID,
			Metadata: auditlog.AccountStatusChangeMetadata{
				OldStatus: "SUSPENDED",
				NewStatus: "ACTIVE",
			},
		}

		m.db.EXPECT().GetAccountByID(gomock.Any(), accountID).
			Return(account, nil)
		m.db.EXPECT().UpdateAccountStatus(gomock.Any(), models.UpdateAccountStatusParams{
			ID:     accountID,
			Status: "ACTIVE",
		}).Return(nil)

		m.auditLog.EXPECT().Submit(gomock.Any(), expectedEventLog).
			Return(nil)

		err := m.service.ActivateAccount(context.TODO(), OperationParams{
			UserID:    userID,
			AccountID: accountID,
		})
		assert.NoError(t, err)
	})

	t.Run("activate account - fails when account not found", func(t *testing.T) {
		m := mockAccountService(t)
		accountID, userID := uuid.New(), uuid.New()
		account := models.GetAccountByIDRow{}
		m.db.EXPECT().GetAccountByID(gomock.Any(), accountID).
			Return(account, sql.ErrNoRows)

		err := m.service.ActivateAccount(context.TODO(), OperationParams{
			UserID:    userID,
			AccountID: accountID,
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "account not found")
	})

	t.Run("activate account - return nil error when account is already active", func(t *testing.T) {
		m := mockAccountService(t)
		accountID := uuid.New()
		userID := uuid.New()
		account := models.GetAccountByIDRow{
			ID:     accountID,
			UserID: userID,
			Status: "ACTIVE",
		}

		m.db.EXPECT().GetAccountByID(gomock.Any(), accountID).
			Return(account, nil)

		err := m.service.ActivateAccount(context.TODO(), OperationParams{
			UserID:    userID,
			AccountID: accountID,
		})
		assert.NoError(t, err)
	})
}

func TestService_CloseAccount(t *testing.T) {
	t.Run("successfully closes account", func(t *testing.T) {
		m := mockAccountService(t)
		accountID := uuid.New()
		userID := uuid.New()
		account := models.GetAccountByIDRow{
			ID:     accountID,
			UserID: userID,
			Status: "ACTIVE",
		}
		expectedEventLog := auditlog.Event{
			Action:    auditlog.ActionAccountStatusChange,
			UserID:    userID,
			AccountID: accountID,
			Metadata: auditlog.AccountStatusChangeMetadata{
				OldStatus: "ACTIVE",
				NewStatus: "CLOSED",
			},
		}

		m.db.EXPECT().GetAccountByID(gomock.Any(), accountID).
			Return(account, nil)
		m.db.EXPECT().UpdateAccountStatus(gomock.Any(), models.UpdateAccountStatusParams{
			ID:     accountID,
			Status: "CLOSED",
		}).Return(nil)
		m.auditLog.EXPECT().Submit(gomock.Any(), expectedEventLog).
			Return(nil)

		err := m.service.CloseAccount(context.TODO(), OperationParams{
			UserID:    userID,
			AccountID: accountID,
		})
		assert.NoError(t, err)
	})

	t.Run("close account - fails when account not found", func(t *testing.T) {
		m := mockAccountService(t)
		accountID, userID := uuid.New(), uuid.New()

		m.db.EXPECT().GetAccountByID(gomock.Any(), accountID).
			Return(models.GetAccountByIDRow{}, sql.ErrNoRows)

		err := m.service.CloseAccount(context.TODO(), OperationParams{
			UserID:    userID,
			AccountID: accountID,
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "account not found")
	})

	t.Run("fails when account is already closed", func(t *testing.T) {
		m := mockAccountService(t)
		accountID := uuid.New()
		userID := uuid.New()
		account := models.GetAccountByIDRow{
			ID:            accountID,
			UserID:        userID,
			AccountNumber: "1234567890",
			Status:        "CLOSED",
		}

		m.db.EXPECT().GetAccountByID(gomock.Any(), accountID).
			Return(account, nil)

		err := m.service.CloseAccount(context.TODO(), OperationParams{
			UserID:    userID,
			AccountID: accountID,
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "account is already closed")
	})
}

func TestService_GetAccountStatusHistory(t *testing.T) {
	t.Run("successfully gets account status history", func(t *testing.T) {
		m := mockAccountService(t)
		accountID := uuid.New()

		historyRows := []models.GetAccountStatusHistoryRow{
			{
				AccountID:     accountID,
				OldStatus:     "ACTIVE",
				NewStatus:     "SUSPENDED",
				ActionBy:      "system",
				CurrentStatus: "SUSPENDED",
			},
			{
				AccountID:     accountID,
				OldStatus:     "ACTIVE",
				NewStatus:     "SUSPENDED",
				ActionBy:      "system",
				CurrentStatus: "SUSPENDED",
			},
		}
		expectedHistory := []ChangeHistory{
			{
				AccountID:     accountID,
				OldStatus:     "ACTIVE",
				NewStatus:     "SUSPENDED",
				ActionBy:      "system",
				CurrentStatus: "SUSPENDED",
			},
			{
				AccountID:     accountID,
				OldStatus:     "ACTIVE",
				NewStatus:     "SUSPENDED",
				ActionBy:      "system",
				CurrentStatus: "SUSPENDED",
			},
		}

		m.db.EXPECT().GetAccountStatusHistory(gomock.Any(), uuid.NullUUID{
			UUID:  accountID,
			Valid: true,
		}).Return(historyRows, nil)

		history, err := m.service.GetAccountStatusHistory(context.TODO(), accountID)

		assert.NoError(t, err)
		assert.Len(t, history, 2)
		assert.Equal(t, expectedHistory, history)

	})

	t.Run("successfully returns empty history", func(t *testing.T) {
		m := mockAccountService(t)
		accountID := uuid.New()

		m.db.EXPECT().GetAccountStatusHistory(gomock.Any(), uuid.NullUUID{
			UUID:  accountID,
			Valid: true,
		}).Return([]models.GetAccountStatusHistoryRow{}, nil)

		history, err := m.service.GetAccountStatusHistory(context.TODO(), accountID)

		assert.NoError(t, err)
		assert.Empty(t, history)
	})

	t.Run("returns not found error when account doesn't exist", func(t *testing.T) {
		m := mockAccountService(t)
		accountID := uuid.New()

		m.db.EXPECT().GetAccountStatusHistory(gomock.Any(), uuid.NullUUID{
			UUID:  accountID,
			Valid: true,
		}).Return(nil, sql.ErrNoRows)

		history, err := m.service.GetAccountStatusHistory(context.TODO(), accountID)

		assert.Error(t, err)
		assert.Nil(t, history)
		assert.Contains(t, err.Error(), "account not found")
	})
}

type accountServiceMocker struct {
	db              *databasemocks.MockQuerier
	generator       *generatormocks.MockTokenGenerator
	numberGenerator *generatormocks.MockNumberGenerator
	passwordHasher  *passwordhashermocks.MockHasher
	txService       *transaction.MockService
	auditLog        *auditlog.MockService

	service Service
}

func mockAccountService(t *testing.T) *accountServiceMocker {
	ctrl := gomock.NewController(t)
	mockDB := databasemocks.NewMockQuerier(ctrl)
	mockGenerator := generatormocks.NewMockTokenGenerator(ctrl)
	mockNumberGen := generatormocks.NewMockNumberGenerator(ctrl)
	passwordHasher := passwordhashermocks.NewMockHasher(ctrl)
	txServiceMock := transaction.NewMockService(ctrl)
	auditLogMock := auditlog.NewMockService(ctrl)

	generator.DefaultNumberGenerator = mockNumberGen
	password.DefaultPasswordHasher = passwordHasher

	svc := NewService(mockDB, auditLogMock, txServiceMock, mockGenerator)
	return &accountServiceMocker{
		db:              mockDB,
		generator:       mockGenerator,
		txService:       txServiceMock,
		auditLog:        auditLogMock,
		numberGenerator: mockNumberGen,
		passwordHasher:  passwordHasher,
		service:         svc,
	}
}
