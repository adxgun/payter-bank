//go:generate mockgen -source=service.go -destination=service_mock.go -package=account
package account

import (
	"context"
	"database/sql"
	"errors"
	"go.uber.org/zap"
	"payter-bank/internal/database/models"
	platformerrors "payter-bank/internal/errors"
	"payter-bank/internal/logger"
	"payter-bank/pkg/generator"
)

type Service interface {
	InitialiseAdmin(ctx context.Context, email, password string) error
}

type service struct {
	db             models.Querier
	tokenGenerator generator.TokenGenerator
}

func NewAccountService(db models.Querier, tokenGenerator generator.TokenGenerator) Service {
	return &service{
		db:             db,
		tokenGenerator: tokenGenerator,
	}
}

func (s service) InitialiseAdmin(ctx context.Context, email, password string) error {
	ctx = logger.With(ctx,
		zap.String(logger.FunctionName, "InitialiseAdmin"),
		zap.String("email", email))
	user, err := s.db.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Error(ctx, "failed to check if user exists", zap.Error(err))
			return platformerrors.ErrInternal
		}
	}

	if user.Email != "" {
		return platformerrors.MakeApiError(400, "admin user already exists")
	}

	newUser, err := s.db.SaveUser(ctx, models.SaveUserParams{
		Email:    email,
		Password: password,
	})
	if err != nil {
		logger.Error(ctx, "failed to save user", zap.Error(err))
		return platformerrors.ErrInternal
	}

	_, err = s.db.SaveAccount(ctx, models.SaveAccountParams{
		UserID:        newUser.ID,
		FirstName:     "Admin",
		LastName:      "Admin",
		AccountType:   models.AccountTypeADMIN,
		Status:        models.StatusACTIVE,
		AccountNumber: generator.DefaultNumberGenerator.Generate(),
	})
	if err != nil {
		logger.Error(ctx, "failed to save account", zap.Error(err))
		return platformerrors.ErrInternal
	}

	return nil
}
