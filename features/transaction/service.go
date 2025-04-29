//go:generate mockgen -source=service.go -destination=service_mock.go -package=transaction

package transaction

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"net/http"
	"payter-bank/features/auditlog"
	"payter-bank/internal/database/models"
	platformerrors "payter-bank/internal/errors"
	"payter-bank/internal/logger"
	"payter-bank/internal/pkg/generator"
)

type Service interface {
	CreditAccount(ctx context.Context, req AccountTransactionParams) (*Response, error)
	DebitAccount(ctx context.Context, req AccountTransactionParams) (*Response, error)
	Transfer(ctx context.Context, req AccountTransactionParams) (*Response, error)
	GetTransactionHistory(ctx context.Context, accountID uuid.UUID) ([]Transaction, error)
	GetAccountBalance(ctx context.Context, accountID uuid.UUID) (Balance, error)
}

type transactionService struct {
	db       models.Querier
	auditLog auditlog.Service
}

func NewService(db models.Querier, auditLog auditlog.Service) Service {
	return &transactionService{
		db:       db,
		auditLog: auditLog,
	}
}

func (t *transactionService) CreditAccount(ctx context.Context, req AccountTransactionParams) (*Response, error) {
	ctx = logger.With(ctx,
		zap.String(logger.FunctionName, "CreditAccount"),
		zap.Any(logger.RequestFields, req))

	if req.FromAccountID == req.ToAccountID {
		return nil, platformerrors.MakeApiError(http.StatusBadRequest, "cannot credit the same account")
	}

	fromAccount, err := t.db.GetAccountByID(ctx, req.FromAccountID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, platformerrors.MakeApiError(http.StatusNotFound, "account not found")
		}
		logger.Error(ctx, "failed to get account by ID", zap.Error(err))
		return nil, platformerrors.ErrInternal
	}

	toAccount, err := t.db.GetAccountByID(ctx, req.ToAccountID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, platformerrors.MakeApiError(http.StatusNotFound, "account not found")
		}
		logger.Error(ctx, "failed to get account by ID", zap.Error(err))
		return nil, platformerrors.ErrInternal
	}

	balance, err := t.db.GetAccountBalance(ctx, fromAccount.ID)
	if err != nil {
		logger.Error(ctx, "failed to get account balance", zap.Error(err))
		return nil, err
	}

	if fromAccount.AccountType != models.AccountTypeEXTERNAL && int64(balance.Balance) < req.AmountUnit() {
		return nil, platformerrors.MakeApiError(http.StatusPreconditionFailed, "insufficient funds")
	}

	if fromAccount.Currency != toAccount.Currency {
		return nil, platformerrors.MakeApiError(http.StatusPreconditionFailed, fmt.Sprintf("you cannot credit %s account with %s account", fromAccount.Currency, toAccount.Currency))
	}

	transaction, err := t.db.SaveTransaction(ctx, models.SaveTransactionParams{
		FromAccountID:   fromAccount.ID,
		ToAccountID:     toAccount.ID,
		Amount:          req.AmountUnit(),
		ReferenceNumber: generator.DefaultNumberGenerator.Generate(),
		Description: sql.NullString{
			String: req.Narration,
			Valid:  req.Narration != "",
		},
		Status:   "COMPLETED",
		Currency: string(fromAccount.Currency),
	})
	if err != nil {
		logger.Error(ctx, "failed to save transaction", zap.Error(err))
		return nil, platformerrors.ErrInternal
	}

	t.updateBalance(ctx, fromAccount.ID)
	t.updateBalance(ctx, toAccount.ID)

	auditEvent := auditlog.NewEvent(auditlog.ActionAccountCredit, req.UserID, toAccount.ID, transaction)
	err = t.auditLog.Submit(ctx, auditEvent)
	if err != nil {
		logger.Error(ctx, "failed to submit audit event", zap.Error(err))
	}

	return &Response{
		TransactionID: transaction.ID,
	}, nil
}

func (t *transactionService) DebitAccount(ctx context.Context, req AccountTransactionParams) (*Response, error) {
	ctx = logger.With(ctx,
		zap.String(logger.FunctionName, "DebitAccount"),
		zap.Any(logger.RequestFields, req))

	if req.FromAccountID == req.ToAccountID {
		return nil, platformerrors.MakeApiError(http.StatusBadRequest, "cannot debit the same account")
	}

	fromAccount, err := t.db.GetAccountByID(ctx, req.FromAccountID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, platformerrors.MakeApiError(http.StatusNotFound, "account not found")
		}
		logger.Error(ctx, "failed to get account by ID", zap.Error(err))
		return nil, platformerrors.ErrInternal
	}

	toAccount, err := t.db.GetAccountByID(ctx, req.ToAccountID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, platformerrors.MakeApiError(http.StatusNotFound, "account not found")
		}
		logger.Error(ctx, "failed to get account by ID", zap.Error(err))
		return nil, platformerrors.ErrInternal
	}

	balance, err := t.db.GetAccountBalance(ctx, fromAccount.ID)
	if err != nil {
		logger.Error(ctx, "failed to get account balance", zap.Error(err))
		return nil, err
	}

	if int64(balance.Balance) < req.AmountUnit() {
		return nil, platformerrors.MakeApiError(http.StatusPreconditionFailed, "insufficient funds")
	}

	if fromAccount.Currency != toAccount.Currency {
		return nil, platformerrors.MakeApiError(http.StatusPreconditionFailed, fmt.Sprintf("you cannot debit %s account with %s account", fromAccount.Currency, toAccount.Currency))
	}

	transaction, err := t.db.SaveTransaction(ctx, models.SaveTransactionParams{
		FromAccountID:   fromAccount.ID,
		ToAccountID:     toAccount.ID,
		Amount:          req.AmountUnit(),
		ReferenceNumber: generator.DefaultNumberGenerator.Generate(),
		Description: sql.NullString{
			String: req.Narration,
			Valid:  req.Narration != "",
		},
		Status:   "COMPLETED",
		Currency: string(fromAccount.Currency),
	})
	if err != nil {
		logger.Error(ctx, "failed to save transaction", zap.Error(err))
		return nil, platformerrors.ErrInternal
	}

	t.updateBalance(ctx, fromAccount.ID)
	t.updateBalance(ctx, toAccount.ID)

	auditEvent := auditlog.NewEvent(auditlog.ActionAccountDebit, req.UserID, fromAccount.ID, transaction)
	err = t.auditLog.Submit(ctx, auditEvent)
	if err != nil {
		logger.Error(ctx, "failed to submit audit event", zap.Error(err))
	}

	return &Response{
		TransactionID: transaction.ID,
	}, nil
}

func (t *transactionService) Transfer(ctx context.Context, req AccountTransactionParams) (*Response, error) {
	return t.DebitAccount(ctx, req)
}

func (t *transactionService) GetTransactionHistory(ctx context.Context, accountID uuid.UUID) ([]Transaction, error) {
	rows, err := t.db.GetTransactionsByAccountID(ctx, accountID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, platformerrors.MakeApiError(http.StatusNotFound, "account not found")
		}
		logger.Error(ctx, "failed to get transaction history", zap.Error(err))
		return nil, err
	}

	transactions := make([]Transaction, 0, len(rows))
	for _, row := range rows {
		transactions = append(transactions, TransactionFromRow(row))
	}
	return transactions, nil
}

func (t *transactionService) GetAccountBalance(ctx context.Context, accountID uuid.UUID) (Balance, error) {
	bal, err := t.db.GetAccountBalance(ctx, accountID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Balance{}, platformerrors.MakeApiError(http.StatusNotFound, "account not found")
		}
		logger.Error(ctx, "failed to get account balance", zap.Error(err))
		return Balance{}, err
	}

	return BalanceFromQueryResult(bal), nil
}

func (t *transactionService) updateBalance(ctx context.Context, accountID uuid.UUID) {
	err := t.db.UpdateBalance(ctx, accountID)
	if err != nil {
		logger.Error(ctx, "failed to update account balance", zap.Error(err))
	}
}
