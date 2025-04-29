//go:generate mockgen -source=service.go -destination=service_mock.go -package=interestrate

package interestrate

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"payter-bank/features/auditlog"
	"payter-bank/internal/config"
	"payter-bank/internal/database/models"
	platformerrors "payter-bank/internal/errors"
	"payter-bank/internal/logger"
	"payter-bank/internal/pkg/generator"
	"syscall"
	"time"
)

var (
	interestRateApplicationJobID = uuid.MustParse("00000000-0000-0000-0000-000000000001")
)

type Service interface {
	CreateInterestRate(ctx context.Context, param CreateInterestRateParam) (*Response, error)
	UpdateRate(ctx context.Context, param UpdateRateParam) (*Response, error)
	UpdateCalculationFrequency(ctx context.Context, param UpdateCalculationFrequencyParam) (*Response, error)
	GetCurrentRate(ctx context.Context) (*models.InterestRate, error)
	ApplyRates(ctx context.Context) error
	Start(ctx context.Context) error
}

type Runner interface {
	Start(ctx context.Context) error
}

func NewRunner(db models.Querier, cfg config.AppConfig) Runner {
	return &service{
		db:  db,
		cfg: cfg,
	}
}

type service struct {
	db       models.Querier
	auditLog auditlog.Service
	cfg      config.AppConfig
	runner   Runner
}

func NewService(db models.Querier, cfg config.AppConfig, auditLog auditlog.Service, runner Runner) Service {
	return &service{
		db:       db,
		cfg:      cfg,
		auditLog: auditLog,
		runner:   runner,
	}
}

func (s *service) CreateInterestRate(ctx context.Context, param CreateInterestRateParam) (*Response, error) {
	ctx = logger.With(ctx,
		zap.String(logger.FunctionName, "CreateInterestRate"),
		zap.Any(logger.RequestFields, param))

	existingRates, err := s.db.GetInterestRates(ctx)
	if err != nil {
		logger.Error(ctx, "failed to get existing interest rates", zap.Error(err))
		return nil, err
	}

	if len(existingRates) > 0 {
		rate := existingRates[0]
		return &Response{InterestRateID: rate.ID}, nil
	}

	newRate, err := s.db.SaveInterestRate(ctx, models.SaveInterestRateParams{
		Rate:                 int64(param.Rate * 100),
		CalculationFrequency: param.CalculationFrequency,
	})
	if err != nil {
		logger.Error(ctx, "failed to save interest rate", zap.Error(err))
		return nil, err
	}

	auditEvent := auditlog.NewEvent(
		auditlog.ActionInterestRateChange, param.UserID, uuid.Nil,
		auditlog.InterestRateChangeMetadata{
			NewRate:                 int64(param.Rate * 100),
			NewCalculationFrequency: param.CalculationFrequency,
		},
	)
	err = s.auditLog.Submit(ctx, auditEvent)
	if err != nil {
		logger.Warn(ctx, "failed to submit audit log", zap.Error(err))
	}

	// restart the scheduler to apply the new rate
	_ = s.runner.Start(ctx)
	return &Response{InterestRateID: newRate.ID}, nil
}

func (s *service) UpdateRate(ctx context.Context, param UpdateRateParam) (*Response, error) {
	ctx = logger.With(ctx,
		zap.String(logger.FunctionName, "UpdateRate"),
		zap.Any(logger.RequestFields, param))

	rate, err := s.GetCurrentRate(ctx)
	if err != nil {
		return nil, err
	}

	err = s.db.UpdateRate(ctx, models.UpdateRateParams{
		ID:   rate.ID,
		Rate: int64(param.Rate * 100),
	})
	if err != nil {
		logger.Error(ctx, "failed to update interest rate", zap.Error(err))
		return nil, platformerrors.ErrInternal
	}
	auditEvent := auditlog.NewEvent(
		auditlog.ActionInterestRateChange, param.UserID, uuid.Nil,
		auditlog.InterestRateChangeMetadata{
			OldRate: rate.Rate,
			NewRate: int64(param.Rate * 100),
		})
	err = s.auditLog.Submit(ctx, auditEvent)
	if err != nil {
		logger.Warn(ctx, "failed to submit audit log", zap.Error(err))
	}

	// restart the scheduler to apply the new rate
	_ = s.runner.Start(ctx)
	return &Response{InterestRateID: rate.ID}, nil
}

func (s *service) UpdateCalculationFrequency(ctx context.Context, param UpdateCalculationFrequencyParam) (*Response, error) {
	ctx = logger.With(ctx,
		zap.String(logger.FunctionName, "UpdateCalculationFrequency"),
		zap.Any(logger.RequestFields, param))

	rate, err := s.GetCurrentRate(ctx)
	if err != nil {
		return nil, err
	}

	if rate.CalculationFrequency == param.CalculationFrequency {
		return &Response{InterestRateID: rate.ID}, nil
	}

	err = s.db.UpdateCalculationFrequency(ctx, models.UpdateCalculationFrequencyParams{
		ID:                   rate.ID,
		CalculationFrequency: param.CalculationFrequency,
	})
	if err != nil {
		logger.Error(ctx, "failed to update interest rate", zap.Error(err))
		return nil, platformerrors.ErrInternal
	}
	auditEvent := auditlog.NewEvent(auditlog.ActionInterestRateChange, param.UserID, uuid.Nil,
		auditlog.InterestRateChangeMetadata{
			OldCalculationFrequency: rate.CalculationFrequency,
			NewCalculationFrequency: param.CalculationFrequency,
		})
	err = s.auditLog.Submit(ctx, auditEvent)
	if err != nil {
		logger.Warn(ctx, "failed to submit audit log", zap.Error(err))
	}

	// restart the scheduler to apply the new rate
	_ = s.runner.Start(ctx)
	return &Response{InterestRateID: rate.ID}, nil
}

func (s *service) ApplyRates(ctx context.Context) error {
	ctx = logger.With(ctx,
		zap.String(logger.FunctionName, "ApplyRates"))

	rate, err := s.GetCurrentRate(ctx)
	if err != nil {
		logger.Error(ctx, "failed to get current rate", zap.Error(err))
		return err
	}

	accounts, err := s.db.GetAllActiveAccounts(ctx)
	if err != nil {
		logger.Error(ctx, "failed to get all active accounts", zap.Error(err))
		return platformerrors.ErrInternal
	}

	for _, account := range accounts {
		balance, err := s.db.GetAccountBalance(ctx, account.AccountID)
		if err != nil {
			logger.Error(ctx, "failed to get account balance", zap.Error(err))
			continue
		}

		if balance.Balance > 0 {
			gain := float64(balance.Balance) * (float64(rate.Rate) / 10000)
			txn := models.SaveTransactionParams{
				FromAccountID:   s.cfg.InterestRateAccountID,
				ToAccountID:     account.AccountID,
				Amount:          int64(gain),
				ReferenceNumber: generator.DefaultNumberGenerator.Generate(),
				Description: sql.NullString{
					String: fmt.Sprintf("Interest gained on %s", time.Now().Format(time.DateOnly)),
				},
				Status:   "COMPLETED",
				Currency: string(account.Currency),
			}
			newTxn, err := s.db.SaveTransaction(ctx, txn)
			if err != nil {
				logger.Error(ctx, "failed to save transaction", zap.Error(err))
				continue
			}
			logger.Info(ctx, "interest applied successfully",
				zap.String("account_id", account.AccountID.String()),
				zap.String("transaction_id", newTxn.ID.String()))
		}
	}

	return nil
}

func (s *service) GetCurrentRate(ctx context.Context) (*models.InterestRate, error) {
	existingRates, err := s.db.GetInterestRates(ctx)
	if err != nil {
		return nil, err
	}

	if len(existingRates) == 0 {
		return nil, platformerrors.MakeApiError(http.StatusPreconditionFailed, "interest rate has not been initialized")
	}

	rate := existingRates[0]
	return &rate, nil
}

func (s *service) Start(ctx context.Context) error {
	// create a context that will be cancelled when the application is shutting down
	ctx, _ = signal.NotifyContext(ctx, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGINT)
	ctx = logger.With(ctx,
		zap.String(logger.FunctionName, "Start"))

	rate, err := s.GetCurrentRate(ctx)
	if err != nil {
		return err
	}

	scheduler, err := gocron.NewScheduler()
	if err != nil {
		return err
	}

	cronExpression, err := s.frequencyToCron(Frequency(rate.CalculationFrequency))
	if err != nil {
		return err
	}

	err = scheduler.RemoveJob(interestRateApplicationJobID) // tries to remove the job if it exists
	if err != nil {
		logger.Warn(ctx, "failed to remove existing job", zap.Error(err))
	}

	job, err := scheduler.NewJob(
		gocron.CronJob(cronExpression, false),
		gocron.NewTask(s.ApplyRates, ctx),
		gocron.WithIdentifier(interestRateApplicationJobID))
	if err != nil {
		return err
	}

	logger.Info(ctx, "interest rate job queued",
		zap.Any("ID", job.ID()))

	scheduler.Start()
	for {
		select {
		case <-ctx.Done():
			logger.Info(ctx, "shutting down scheduler")
			if err := scheduler.Shutdown(); err != nil {
				logger.Error(ctx, "failed to shutdown scheduler", zap.Error(err))
			}
			return nil
		default:
		}
	}
}

func (s *service) frequencyToCron(frequency Frequency) (string, error) {
	switch frequency {
	case Hourly:
		return "0 * * * *", nil
	case Daily:
		return "0 0 * * *", nil
	case Weekly:
		return "0 0 * * 0", nil
	case Monthly:
		return "0 0 1 * *", nil
	case Yearly:
		return "0 0 1 1 *", nil
	default:
		return "", fmt.Errorf("unknown frequency: %s", frequency)
	}
}
