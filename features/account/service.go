//go:generate mockgen -source=service.go -destination=service_mock.go -package=account
package account

import (
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"payter-bank/features/auditlog"
	"payter-bank/features/transaction"
	"payter-bank/internal/database/models"
	platformerrors "payter-bank/internal/errors"
	"payter-bank/internal/logger"
	"payter-bank/internal/pkg/generator"
	"payter-bank/internal/pkg/password"
)

type Service interface {
	InitialiseAdmin(ctx context.Context, email, password string) error
	CreateUser(ctx context.Context, param CreateUserParams) (CreateUserResponse, error)
	CreateAccount(ctx context.Context, param CreateAccountParams) (Profile, error)
	AuthenticateAccount(ctx context.Context, param AuthenticateAccountParams) (AccessToken, error)
	GetProfile(ctx context.Context, userID uuid.UUID) (Profile, error)
	SuspendAccount(ctx context.Context, param OperationParams) error
	ActivateAccount(ctx context.Context, param OperationParams) error
	CloseAccount(ctx context.Context, param OperationParams) error
	GetAccountStatusHistory(ctx context.Context, accountID uuid.UUID) ([]ChangeHistory, error)
	GetAllAccounts(ctx context.Context) ([]Account, error)
	GetAccountsStats(ctx context.Context) (models.GetAccountStatsRow, error)
	GetAccountDetails(ctx context.Context, id uuid.UUID) (Account, error)
}

type service struct {
	db                 models.Querier
	auditLog           auditlog.Service
	transactionService transaction.Service
	tokenGenerator     generator.TokenGenerator
}

func NewService(
	db models.Querier,
	auditLog auditlog.Service,
	txService transaction.Service,
	tokenGenerator generator.TokenGenerator) Service {
	return &service{
		db:                 db,
		auditLog:           auditLog,
		transactionService: txService,
		tokenGenerator:     tokenGenerator,
	}
}

func (s service) InitialiseAdmin(ctx context.Context, email, pwd string) error {
	ctx = logger.With(ctx,
		zap.String(logger.FunctionName, "InitialiseAdmin"),
		zap.String("email", email))
	user, err := s.db.GetUserByEmail(ctx, email)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			logger.Error(ctx, "failed to check if user exists", zap.Error(err))
			return platformerrors.ErrInternal
		}
	}

	if user.Email != "" {
		logger.Info(ctx, "admin user has already been created", zap.String("email", email))
		return nil
	}

	newUser, err := s.db.SaveUser(ctx, models.SaveUserParams{
		Email:     email,
		Password:  password.DefaultPasswordHasher.Hash(pwd),
		FirstName: "Admin",
		LastName:  "Admin",
		UserType:  models.UserTypeADMIN,
	})
	if err != nil {
		logger.Error(ctx, "failed to save user", zap.Error(err))
		return platformerrors.ErrInternal
	}

	_, err = s.db.SaveAccount(ctx, models.SaveAccountParams{
		UserID:        newUser.ID,
		AccountType:   models.AccountTypeCURRENT,
		Status:        models.StatusACTIVE,
		AccountNumber: generator.DefaultNumberGenerator.Generate(),
		Currency:      models.CurrencyGBP,
	})
	if err != nil {
		logger.Error(ctx, "failed to save account", zap.Error(err))
		return platformerrors.ErrInternal
	}

	return nil
}

func (s service) CreateUser(ctx context.Context, param CreateUserParams) (CreateUserResponse, error) {
	user, err := s.db.GetUserByEmail(ctx, param.Email)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			logger.Error(ctx, "failed to check if user exists", zap.Error(err))
			return CreateUserResponse{}, platformerrors.ErrInternal
		}
	}

	if user.Email != "" {
		return CreateUserResponse{}, platformerrors.MakeApiError(400, "user already exists")
	}

	newUser, err := s.db.SaveUser(ctx, models.SaveUserParams{
		Email:     param.Email,
		Password:  password.DefaultPasswordHasher.Hash(param.Password),
		FirstName: param.FirstName,
		LastName:  param.LastName,
		UserType:  models.UserType(param.UserType),
	})
	if err != nil {
		logger.Error(ctx, "failed to save user", zap.Error(err))
		return CreateUserResponse{}, platformerrors.ErrInternal
	}
	return CreateUserResponse{
		UserID: newUser.ID,
	}, nil
}

func (s service) CreateAccount(ctx context.Context, param CreateAccountParams) (Profile, error) {
	ctx = logger.With(ctx,
		zap.String(logger.FunctionName, "CreateAccount"),
		zap.Any(logger.RequestFields, param))

	user, err := s.db.GetUserByID(ctx, param.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Profile{}, platformerrors.MakeApiError(404, "user not found")
		}
		logger.Error(ctx, "failed to get user by id", zap.Error(err))
		return Profile{}, platformerrors.ErrInternal
	}

	existingAccount, err := s.db.GetAccountByCurrency(ctx, models.GetAccountByCurrencyParams{
		Currency: models.Currency(param.Currency),
		UserID:   param.UserID,
	})
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			logger.Error(ctx, "failed to get account by currency", zap.Error(err))
			return Profile{}, platformerrors.ErrInternal
		}
	}

	if existingAccount.ID != uuid.Nil {
		return Profile{}, platformerrors.MakeApiError(400, "account already exists")
	}

	account := models.SaveAccountParams{
		UserID:        user.ID,
		AccountType:   models.AccountTypeCURRENT,
		Status:        models.StatusACTIVE,
		AccountNumber: generator.DefaultNumberGenerator.Generate(),
		Currency:      models.Currency(param.Currency),
	}

	newAccount, err := s.db.SaveAccount(ctx, account)
	if err != nil {
		logger.Error(ctx, "failed to save account", zap.Error(err))
		return Profile{}, platformerrors.ErrInternal
	}

	err = s.auditLog.Submit(ctx, auditlog.NewEvent(auditlog.ActionCreateAccount, param.AdminUserID, newAccount.ID, param))
	if err != nil {
		logger.Error(ctx, "failed to queue audit log", zap.Error(err))
	}

	if param.InitialDeposit > 0 {
		txParam := transaction.AccountTransactionParams{
			FromAccountID: uuid.Nil,
			ToAccountID:   newAccount.ID,
			Amount:        param.InitialDeposit,
			Narration:     "Initial deposit",
			UserID:        param.UserID,
		}
		_, err = s.transactionService.CreditAccount(ctx, txParam)
		if err != nil {
			logger.Error(ctx, "failed to credit account", zap.Error(err))
			return Profile{}, platformerrors.ErrInternal
		}
	}

	p, err := s.db.GetProfileByUserID(ctx, user.ID)
	if err != nil {
		logger.Error(ctx, "failed to get profile by user id", zap.Error(err))
		return Profile{}, platformerrors.ErrInternal
	}
	return ProfileFromQueryResult(p), nil
}

func (s service) AuthenticateAccount(ctx context.Context, param AuthenticateAccountParams) (AccessToken, error) {
	ctx = logger.With(ctx,
		zap.String(logger.FunctionName, "AuthenticateAccount"),
		zap.Any(logger.RequestFields, param.Email))

	user, err := s.db.GetUserByEmail(ctx, param.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return AccessToken{}, platformerrors.MakeApiError(401, "invalid login credentials")
		}
		logger.Error(ctx, "failed to get user by email", zap.Error(err))
		return AccessToken{}, platformerrors.ErrInternal
	}

	if !password.DefaultPasswordHasher.Validate(user.Password, param.Password) {
		return AccessToken{}, platformerrors.MakeApiError(401, "invalid login credentials")
	}

	profile, err := s.db.GetProfileByUserID(ctx, user.ID)
	if err != nil {
		logger.Error(ctx, "failed to get profile by user id", zap.Error(err))
		return AccessToken{}, platformerrors.ErrInternal
	}

	tokenData := generator.TokenData{
		UserID:    profile.UserID,
		AccountID: profile.AccountID,
	}
	token, err := s.tokenGenerator.Generate(tokenData)
	if err != nil {
		logger.Error(ctx, "failed to generate token", zap.Error(err))
		return AccessToken{}, platformerrors.ErrInternal
	}
	return AccessToken{
		Token: token,
	}, nil
}

func (s service) GetProfile(ctx context.Context, userID uuid.UUID) (Profile, error) {
	ctx = logger.With(ctx,
		zap.String(logger.FunctionName, "GetProfile"),
		zap.Any(logger.RequestFields, userID))

	profile, err := s.db.GetProfileByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Profile{}, platformerrors.MakeApiError(404, "profile not found")
		}
		logger.Error(ctx, "failed to get profile by user id", zap.Error(err))
		return Profile{}, platformerrors.ErrInternal
	}

	return ProfileFromQueryResult(profile), nil
}

func (s service) SuspendAccount(ctx context.Context, param OperationParams) error {
	ctx = logger.With(ctx,
		zap.String(logger.FunctionName, "SuspendAccount"),
		zap.Any(logger.RequestFields, param))
	account, err := s.db.GetAccountByID(ctx, param.AccountID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return platformerrors.MakeApiError(404, "account not found")
		}
		logger.Error(ctx, "failed to get account by id", zap.Error(err))
		return platformerrors.ErrInternal
	}

	if account.Status == models.StatusSUSPENDED {
		return platformerrors.MakeApiError(400, "account already suspended")
	}
	err = s.db.UpdateAccountStatus(ctx, models.UpdateAccountStatusParams{
		ID:     account.ID,
		Status: models.StatusSUSPENDED,
	})
	if err != nil {
		logger.Error(ctx, "failed to suspend account", zap.Error(err))
		return platformerrors.ErrInternal
	}

	auditEvent := auditlog.NewEvent(
		auditlog.ActionAccountStatusChange,
		param.UserID,
		account.ID,
		auditlog.AccountStatusChangeMetadata{OldStatus: string(account.Status), NewStatus: string(models.StatusSUSPENDED)})
	err = s.auditLog.Submit(ctx, auditEvent)
	if err != nil {
		logger.Error(ctx, "failed to queue audit log", zap.Error(err))
	}
	return nil
}

func (s service) ActivateAccount(ctx context.Context, param OperationParams) error {
	ctx = logger.With(ctx,
		zap.String(logger.FunctionName, "ActivateAccount"),
		zap.Any(logger.RequestFields, param))

	account, err := s.db.GetAccountByID(ctx, param.AccountID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return platformerrors.MakeApiError(404, "account not found")
		}
		logger.Error(ctx, "failed to get account by id", zap.Error(err))
		return platformerrors.ErrInternal
	}

	if account.Status == models.StatusACTIVE {
		return nil
	}

	err = s.db.UpdateAccountStatus(ctx, models.UpdateAccountStatusParams{
		ID:     account.ID,
		Status: models.StatusACTIVE,
	})
	if err != nil {
		logger.Error(ctx, "failed to activate account", zap.Error(err))
		return platformerrors.ErrInternal
	}

	auditEvent := auditlog.NewEvent(
		auditlog.ActionAccountStatusChange,
		param.UserID,
		account.ID,
		auditlog.AccountStatusChangeMetadata{OldStatus: string(account.Status), NewStatus: string(models.StatusACTIVE)})
	err = s.auditLog.Submit(ctx, auditEvent)
	if err != nil {
		logger.Error(ctx, "failed to queue audit log", zap.Error(err))
	}
	return nil
}

func (s service) CloseAccount(ctx context.Context, param OperationParams) error {
	ctx = logger.With(ctx,
		zap.String(logger.FunctionName, "CloseAccount"),
		zap.Any(logger.RequestFields, param))

	account, err := s.db.GetAccountByID(ctx, param.AccountID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return platformerrors.MakeApiError(404, "account not found")
		}
		logger.Error(ctx, "failed to get account by id", zap.Error(err))
		return platformerrors.ErrInternal
	}

	if account.Status == models.StatusCLOSED {
		return platformerrors.MakeApiError(400, "account is already closed")
	}

	err = s.db.UpdateAccountStatus(ctx, models.UpdateAccountStatusParams{
		ID:     account.ID,
		Status: models.StatusCLOSED,
	})
	if err != nil {
		logger.Error(ctx, "failed to close account", zap.Error(err))
		return platformerrors.ErrInternal
	}

	auditEvent := auditlog.NewEvent(
		auditlog.ActionAccountStatusChange,
		param.UserID,
		account.ID,
		auditlog.AccountStatusChangeMetadata{OldStatus: string(account.Status), NewStatus: string(models.StatusCLOSED)})
	err = s.auditLog.Submit(ctx, auditEvent)
	if err != nil {
		logger.Error(ctx, "failed to queue audit log", zap.Error(err))
	}

	return nil
}

func (s service) GetAccountStatusHistory(ctx context.Context, accountID uuid.UUID) ([]ChangeHistory, error) {
	rows, err := s.db.GetAccountStatusHistory(ctx, uuid.NullUUID{
		UUID:  accountID,
		Valid: true,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, platformerrors.MakeApiError(404, "account not found")
		}
		logger.Error(ctx, "failed to get account status history", zap.Error(err))
		return nil, platformerrors.ErrInternal
	}

	items := make([]ChangeHistory, 0, len(rows))
	for _, row := range rows {
		items = append(items, ChangeHistoryFromRow(row))
	}
	return items, nil
}

func (s service) GetAllAccounts(ctx context.Context) ([]Account, error) {
	data, err := s.db.GetAllCurrentAccounts(ctx)
	if err != nil {
		logger.Error(ctx, "failed to get all accounts", zap.Error(err))
		return nil, platformerrors.ErrInternal
	}

	accounts := make([]Account, 0, len(data))
	for _, row := range data {
		accounts = append(accounts, AccountFromQuery(row))
	}

	return accounts, nil
}

func (s service) GetAccountsStats(ctx context.Context) (models.GetAccountStatsRow, error) {
	return s.db.GetAccountStats(ctx)
}

func (s service) GetAccountDetails(ctx context.Context, id uuid.UUID) (Account, error) {
	row, err := s.db.GetAccountDetailsByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Account{}, platformerrors.MakeApiError(404, "account not found")
		}
		logger.Error(ctx, "failed to get account details", zap.Error(err))
		return Account{}, platformerrors.ErrInternal
	}
	return AccountFromDetailsRow(row), nil
}
