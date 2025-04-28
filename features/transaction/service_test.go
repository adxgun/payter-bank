package transaction

import (
	"context"
	"database/sql"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"payter-bank/features/auditlog"
	"payter-bank/internal/database/models"
	databasemocks "payter-bank/internal/database/models/mocks"
	platformerrors "payter-bank/internal/errors"
	"payter-bank/internal/pkg/generator"
	generatormocks "payter-bank/internal/pkg/generator/mocks"
	"testing"
)

func TestService_CreditAccount(t *testing.T) {
	t.Run("successfully credits account", func(t *testing.T) {
		m := newTransactionServiceMocker(t)

		req := AccountTransactionParams{
			FromAccountID: uuid.New(),
			ToAccountID:   uuid.New(),
			Amount:        100.50,
			Narration:     "Test credit",
			UserID:        uuid.New(),
		}

		fromAccount := models.GetAccountByIDRow{
			ID:          req.FromAccountID,
			Currency:    models.CurrencyGBP,
			AccountType: models.AccountTypeEXTERNAL,
		}

		toAccount := models.GetAccountByIDRow{
			ID:          req.ToAccountID,
			Currency:    models.CurrencyGBP,
			AccountType: models.AccountTypeCURRENT,
		}

		balance := models.GetAccountBalanceRow{
			AccountID: req.FromAccountID,
			Balance:   20000, // 200.00
		}

		expectedTx := models.Transaction{
			ID:              uuid.New(),
			FromAccountID:   req.FromAccountID,
			ToAccountID:     req.ToAccountID,
			Amount:          req.AmountUnit(),
			ReferenceNumber: "TEST123",
			Status:          "COMPLETED",
			Currency:        string(models.CurrencyGBP),
		}
		expectedSaveTxParams := models.SaveTransactionParams{
			FromAccountID:   req.FromAccountID,
			ToAccountID:     req.ToAccountID,
			Amount:          10050,
			ReferenceNumber: "1234567890",
			Description: sql.NullString{
				String: "Test credit",
				Valid:  true,
			},
			Status:   "COMPLETED",
			Currency: "GBP",
		}

		expectedAuditLogEvent := auditlog.Event{
			Action:    auditlog.ActionAccountCredit,
			UserID:    req.UserID,
			AccountID: req.ToAccountID,
			Metadata:  expectedTx,
		}

		m.numGen.EXPECT().Generate().Return("1234567890")
		m.db.EXPECT().
			GetAccountByID(gomock.Any(), req.FromAccountID).
			Return(fromAccount, nil)

		m.db.EXPECT().
			GetAccountByID(gomock.Any(), req.ToAccountID).
			Return(toAccount, nil)

		m.db.EXPECT().
			GetAccountBalance(gomock.Any(), req.FromAccountID).
			Return(balance, nil)

		m.db.EXPECT().
			SaveTransaction(gomock.Any(), expectedSaveTxParams).
			Return(expectedTx, nil)

		m.auditLog.EXPECT().
			Submit(gomock.Any(), expectedAuditLogEvent).
			Return(nil)

		response, err := m.service.CreditAccount(context.TODO(), req)
		assert.NoError(t, err)
		assert.Equal(t, expectedTx.ID, response.TransactionID)
	})

	t.Run("fails when crediting same account", func(t *testing.T) {
		m := newTransactionServiceMocker(t)
		accountID := uuid.New()
		req := AccountTransactionParams{
			FromAccountID: accountID,
			ToAccountID:   accountID,
			Amount:        100.50,
		}

		_, err := m.service.CreditAccount(context.TODO(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot credit the same account")
	})

	t.Run("fails with insufficient funds", func(t *testing.T) {
		mocker := newTransactionServiceMocker(t)
		req := AccountTransactionParams{
			FromAccountID: uuid.New(),
			ToAccountID:   uuid.New(),
			Amount:        200.00,
		}

		fromAccount := models.GetAccountByIDRow{
			ID:          req.FromAccountID,
			Currency:    models.CurrencyGBP,
			AccountType: models.AccountTypeCURRENT,
		}

		toAccount := models.GetAccountByIDRow{
			ID:          req.ToAccountID,
			Currency:    models.CurrencyGBP,
			AccountType: models.AccountTypeCURRENT,
		}

		balance := models.GetAccountBalanceRow{
			AccountID: req.FromAccountID,
			Balance:   10000, // 100.00
		}

		mocker.db.EXPECT().
			GetAccountByID(gomock.Any(), req.FromAccountID).
			Return(fromAccount, nil)

		mocker.db.EXPECT().
			GetAccountByID(gomock.Any(), req.ToAccountID).
			Return(toAccount, nil)

		mocker.db.EXPECT().
			GetAccountBalance(gomock.Any(), req.FromAccountID).
			Return(balance, nil)

		_, err := mocker.service.CreditAccount(context.TODO(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient funds")
	})
}

func TestService_DebitAccount(t *testing.T) {
	t.Run("successfully debits account", func(t *testing.T) {
		m := newTransactionServiceMocker(t)

		req := AccountTransactionParams{
			FromAccountID: uuid.New(),
			ToAccountID:   uuid.New(),
			Amount:        100.50,
			Narration:     "Test debit",
			UserID:        uuid.New(),
		}

		fromAccount := models.GetAccountByIDRow{
			ID:          req.FromAccountID,
			Currency:    models.CurrencyGBP,
			AccountType: models.AccountTypeCURRENT,
		}

		toAccount := models.GetAccountByIDRow{
			ID:          req.ToAccountID,
			Currency:    models.CurrencyGBP,
			AccountType: models.AccountTypeCURRENT,
		}

		balance := models.GetAccountBalanceRow{
			AccountID: req.FromAccountID,
			Balance:   20000, // 200.00
		}

		expectedTx := models.Transaction{
			ID:              uuid.New(),
			FromAccountID:   req.FromAccountID,
			ToAccountID:     req.ToAccountID,
			Amount:          req.AmountUnit(),
			ReferenceNumber: "TEST123",
			Status:          "COMPLETED",
			Currency:        string(models.CurrencyGBP),
		}

		expectedSaveTxParams := models.SaveTransactionParams{
			FromAccountID:   req.FromAccountID,
			ToAccountID:     req.ToAccountID,
			Amount:          10050,
			ReferenceNumber: "1234567890",
			Description: sql.NullString{
				String: "Test debit",
				Valid:  true,
			},
			Status:   "COMPLETED",
			Currency: "GBP",
		}

		expectedAuditLogEvent := auditlog.Event{
			Action:    auditlog.ActionAccountDebit,
			UserID:    req.UserID,
			AccountID: req.FromAccountID,
			Metadata:  expectedTx,
		}

		m.numGen.EXPECT().Generate().Return("1234567890")
		m.db.EXPECT().
			GetAccountByID(gomock.Any(), req.FromAccountID).
			Return(fromAccount, nil)

		m.db.EXPECT().
			GetAccountByID(gomock.Any(), req.ToAccountID).
			Return(toAccount, nil)

		m.db.EXPECT().
			GetAccountBalance(gomock.Any(), req.FromAccountID).
			Return(balance, nil)

		m.db.EXPECT().
			SaveTransaction(gomock.Any(), expectedSaveTxParams).
			Return(expectedTx, nil)

		m.auditLog.EXPECT().
			Submit(gomock.Any(), expectedAuditLogEvent).
			Return(nil)

		response, err := m.service.DebitAccount(context.TODO(), req)
		assert.NoError(t, err)
		assert.Equal(t, expectedTx.ID, response.TransactionID)
	})

	t.Run("fails when debiting same account", func(t *testing.T) {
		m := newTransactionServiceMocker(t)
		accountID := uuid.New()
		req := AccountTransactionParams{
			FromAccountID: accountID,
			ToAccountID:   accountID,
			Amount:        100.50,
		}

		_, err := m.service.DebitAccount(context.TODO(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot debit the same account")
	})

	t.Run("fails with insufficient funds", func(t *testing.T) {
		m := newTransactionServiceMocker(t)
		req := AccountTransactionParams{
			FromAccountID: uuid.New(),
			ToAccountID:   uuid.New(),
			Amount:        200.00,
		}

		fromAccount := models.GetAccountByIDRow{
			ID:          req.FromAccountID,
			Currency:    models.CurrencyGBP,
			AccountType: models.AccountTypeCURRENT,
		}

		toAccount := models.GetAccountByIDRow{
			ID:          req.ToAccountID,
			Currency:    models.CurrencyGBP,
			AccountType: models.AccountTypeCURRENT,
		}

		balance := models.GetAccountBalanceRow{
			AccountID: req.FromAccountID,
			Balance:   10000, // 100.00
		}

		m.db.EXPECT().
			GetAccountByID(gomock.Any(), req.FromAccountID).
			Return(fromAccount, nil)

		m.db.EXPECT().
			GetAccountByID(gomock.Any(), req.ToAccountID).
			Return(toAccount, nil)

		m.db.EXPECT().
			GetAccountBalance(gomock.Any(), req.FromAccountID).
			Return(balance, nil)

		_, err := m.service.DebitAccount(context.TODO(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient funds")
	})

	t.Run("fails with currency mismatch", func(t *testing.T) {
		m := newTransactionServiceMocker(t)
		req := AccountTransactionParams{
			FromAccountID: uuid.New(),
			ToAccountID:   uuid.New(),
			Amount:        100.50,
		}

		fromAccount := models.GetAccountByIDRow{
			ID:          req.FromAccountID,
			Currency:    models.CurrencyGBP,
			AccountType: models.AccountTypeCURRENT,
		}

		toAccount := models.GetAccountByIDRow{
			ID:          req.ToAccountID,
			Currency:    models.CurrencyEUR,
			AccountType: models.AccountTypeCURRENT,
		}

		balance := models.GetAccountBalanceRow{
			AccountID: req.FromAccountID,
			Balance:   100000, // 1000.00
		}

		m.db.EXPECT().
			GetAccountBalance(gomock.Any(), req.FromAccountID).
			Return(balance, nil)

		m.db.EXPECT().
			GetAccountByID(gomock.Any(), req.FromAccountID).
			Return(fromAccount, nil)

		m.db.EXPECT().
			GetAccountByID(gomock.Any(), req.ToAccountID).
			Return(toAccount, nil)

		_, err := m.service.DebitAccount(context.TODO(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "you cannot debit GBP account with EUR account")
	})

	t.Run("fails when source account not found", func(t *testing.T) {
		m := newTransactionServiceMocker(t)
		req := AccountTransactionParams{
			FromAccountID: uuid.New(),
			ToAccountID:   uuid.New(),
			Amount:        100.50,
		}

		m.db.EXPECT().
			GetAccountByID(gomock.Any(), req.FromAccountID).
			Return(models.GetAccountByIDRow{}, sql.ErrNoRows)

		_, err := m.service.DebitAccount(context.TODO(), req)
		assert.Error(t, err)
	})

	t.Run("fails when destination account not found", func(t *testing.T) {
		m := newTransactionServiceMocker(t)
		req := AccountTransactionParams{
			FromAccountID: uuid.New(),
			ToAccountID:   uuid.New(),
			Amount:        100.50,
		}

		fromAccount := models.GetAccountByIDRow{
			ID:          req.FromAccountID,
			Currency:    models.CurrencyGBP,
			AccountType: models.AccountTypeCURRENT,
		}

		m.db.EXPECT().
			GetAccountByID(gomock.Any(), req.FromAccountID).
			Return(fromAccount, nil)

		m.db.EXPECT().
			GetAccountByID(gomock.Any(), req.ToAccountID).
			Return(models.GetAccountByIDRow{}, sql.ErrNoRows)

		_, err := m.service.DebitAccount(context.TODO(), req)
		assert.Error(t, err)
	})
}

func TestService_GetTransactionHistory(t *testing.T) {
	t.Run("successfully gets transaction history", func(t *testing.T) {
		m := newTransactionServiceMocker(t)
		accountID, accountID1, transactionID1, transactionID2 := uuid.New(), uuid.New(), uuid.New(), uuid.New()

		rows := []models.GetTransactionsByAccountIDRow{
			{
				TransactionID:   transactionID1,
				FromAccountID:   accountID,
				ToAccountID:     accountID1,
				Amount:          10000, // 100.00
				ReferenceNumber: "TRX123456",
				Description:     sql.NullString{String: "Test outgoing transfer", Valid: true},
				Status:          "COMPLETED",
				Currency:        "GBP",
			},
			{
				TransactionID:   transactionID2,
				FromAccountID:   accountID1,
				ToAccountID:     accountID,
				Amount:          20000, // 200.00
				ReferenceNumber: "TRX123457",
				Description:     sql.NullString{String: "Test incoming transfer", Valid: true},
				Status:          "COMPLETED",
				Currency:        "GBP",
			},
		}

		expectedTransactions := []Transaction{
			{
				TransactionID: transactionID1,
				FromAccountID: accountID,
				ToAccountID:   accountID1,
				Amount: Amount{
					Amount:   100.00,
					Currency: "GBP",
				},
				ReferenceNumber: "TRX123456",
				Description:     "Test outgoing transfer",
				Status:          "COMPLETED",
				Currency:        "GBP",
			},
			{
				TransactionID: transactionID2,
				FromAccountID: accountID1,
				ToAccountID:   accountID,
				Amount: Amount{
					Amount:   200.00,
					Currency: "GBP",
				},
				ReferenceNumber: "TRX123457",
				Description:     "Test incoming transfer",
				Status:          "COMPLETED",
				Currency:        "GBP",
			},
		}

		m.db.EXPECT().
			GetTransactionsByAccountID(gomock.Any(), accountID).
			Return(rows, nil)

		transactions, err := m.service.GetTransactionHistory(context.TODO(), accountID)

		assert.NoError(t, err)
		assert.Len(t, transactions, 2)
		assert.Equal(t, expectedTransactions, transactions)
	})

	t.Run("successfully returns empty transaction history", func(t *testing.T) {
		m := newTransactionServiceMocker(t)
		accountID := uuid.New()

		m.db.EXPECT().
			GetTransactionsByAccountID(gomock.Any(), accountID).
			Return([]models.GetTransactionsByAccountIDRow{}, nil)

		transactions, err := m.service.GetTransactionHistory(context.TODO(), accountID)

		assert.NoError(t, err)
		assert.Empty(t, transactions)
	})

	t.Run("returns error when account not found", func(t *testing.T) {
		m := newTransactionServiceMocker(t)
		accountID := uuid.New()

		m.db.EXPECT().
			GetTransactionsByAccountID(gomock.Any(), accountID).
			Return(nil, sql.ErrNoRows)

		transactions, err := m.service.GetTransactionHistory(context.TODO(), accountID)

		assert.Error(t, err)
		assert.Nil(t, transactions)
		assert.Contains(t, err.Error(), "account not found")
	})

	t.Run("returns error on database failure", func(t *testing.T) {
		m := newTransactionServiceMocker(t)
		accountID := uuid.New()

		m.db.EXPECT().
			GetTransactionsByAccountID(gomock.Any(), accountID).
			Return(nil, platformerrors.ErrInternal)

		transactions, err := m.service.GetTransactionHistory(context.TODO(), accountID)

		assert.Error(t, err)
		assert.Nil(t, transactions)
		assert.Contains(t, err.Error(), platformerrors.ErrInternal.Error())
	})
}

func TestService_GetAccountBalance(t *testing.T) {
	t.Run("successfully gets account balance", func(t *testing.T) {
		m := newTransactionServiceMocker(t)
		accountID := uuid.New()

		mockBalance := models.GetAccountBalanceRow{
			AccountID:     accountID,
			Balance:       15000, // 150.00
			AccountNumber: "1234567890",
			AccountType:   models.AccountTypeCURRENT,
			Currency:      models.CurrencyGBP,
		}

		expectedBalance := Balance{
			AccountID:     accountID,
			Balance:       150.00,
			AccountNumber: "1234567890",
			AccountType:   string(models.AccountTypeCURRENT),
			Currency:      string(models.CurrencyGBP),
		}

		m.db.EXPECT().
			GetAccountBalance(gomock.Any(), accountID).
			Return(mockBalance, nil)

		balance, err := m.service.GetAccountBalance(context.TODO(), accountID)

		assert.NoError(t, err)
		assert.Equal(t, expectedBalance, balance)
	})

	t.Run("returns error when account not found", func(t *testing.T) {
		m := newTransactionServiceMocker(t)
		accountID := uuid.New()

		m.db.EXPECT().
			GetAccountBalance(gomock.Any(), accountID).
			Return(models.GetAccountBalanceRow{}, sql.ErrNoRows)

		balance, err := m.service.GetAccountBalance(context.TODO(), accountID)

		assert.Error(t, err)
		assert.Equal(t, Balance{}, balance)
		assert.Contains(t, err.Error(), "account not found")
	})

	t.Run("returns error on database failure", func(t *testing.T) {
		m := newTransactionServiceMocker(t)
		accountID := uuid.New()

		m.db.EXPECT().
			GetAccountBalance(gomock.Any(), accountID).
			Return(models.GetAccountBalanceRow{}, platformerrors.ErrInternal)

		balance, err := m.service.GetAccountBalance(context.TODO(), accountID)

		assert.Error(t, err)
		assert.Equal(t, Balance{}, balance)
		assert.Equal(t, platformerrors.ErrInternal, err)
	})
}

type transactionServiceMocker struct {
	db       *databasemocks.MockQuerier
	auditLog *auditlog.MockService
	numGen   *generatormocks.MockNumberGenerator

	service Service
}

func newTransactionServiceMocker(t *testing.T) *transactionServiceMocker {
	ctrl := gomock.NewController(t)
	db := databasemocks.NewMockQuerier(ctrl)
	auditLog := auditlog.NewMockService(ctrl)
	mockNumberGen := generatormocks.NewMockNumberGenerator(ctrl)

	generator.DefaultNumberGenerator = mockNumberGen

	service := NewService(db, auditLog)
	return &transactionServiceMocker{
		db:       db,
		numGen:   mockNumberGen,
		auditLog: auditLog,
		service:  service,
	}
}
