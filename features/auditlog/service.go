//go:generate mockgen -source=service.go -destination=service_mock.go -package=auditlog

package auditlog

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/sqlc-dev/pqtype"
	"go.uber.org/zap"
	"payter-bank/internal/config"
	"payter-bank/internal/database/models"
	"payter-bank/internal/logger"
)

type Query interface {
	GetAuditLogs(ctx context.Context, accountID uuid.UUID) ([]AuditLog, error)
}

type Service interface {
	Submit(ctx context.Context, event Event) error
	Start(ctx context.Context) error
}

type auditLogProcessor interface {
	ProcessTask(ctx context.Context, task *asynq.Task) error
}

func newProcessor(db models.Querier) auditLogProcessor {
	return &service{
		db: db,
	}
}

type service struct {
	client Client
	cfg    config.Config
	db     models.Querier
}

func NewService(cfg config.Config, client Client, db models.Querier) Service {
	return &service{
		client: client,
		db:     db,
		cfg:    cfg,
	}
}

func NewQueryService(db models.Querier) Query {
	return &service{
		db: db,
	}
}

func (s *service) Start(ctx context.Context) error {
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: s.cfg.Redis.Addr},
		asynq.Config{Concurrency: s.cfg.App.QueueConcurrency})

	mux := asynq.NewServeMux()
	mux.Handle(auditLogTaskName, newProcessor(s.db))

	logger.Info(ctx, "starting audit log service")
	if err := srv.Start(mux); err != nil {
		logger.Error(ctx, "failed to start audit log service", zap.Error(err))
		return err
	}

	for {
		select {
		case <-ctx.Done():
			logger.Info(ctx, "shutting down audit log service")
			srv.Shutdown()
			return nil
		default:
		}
	}
}

func (s *service) Submit(ctx context.Context, event Event) error {
	ctx = logger.With(ctx,
		zap.String(logger.FunctionName, "Submit#AuditLog"),
		zap.Any("event", event))

	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	info, err := s.client.Enqueue(asynq.NewTask(auditLogTaskName, payload))
	if err != nil {
		logger.Error(ctx, "failed to enqueue audit log task", zap.Error(err))
		return err
	}

	logger.Info(ctx, "audit log task enqueued", zap.Any("info", info))
	return nil
}

func (s *service) ProcessTask(ctx context.Context, task *asynq.Task) error {
	ctx = logger.With(ctx,
		zap.String(logger.FunctionName, "Process#AuditLog"),
		zap.Any("task", task.Type()))

	var event Event
	if err := json.Unmarshal(task.Payload(), &event); err != nil {
		logger.Error(ctx, "failed to unmarshal audit log event", zap.Error(err))
		return err
	}

	var (
		metadata []byte
		err      error
	)
	if event.Metadata != nil {
		metadata, err = json.Marshal(event.Metadata)
		if err != nil {
			logger.Error(ctx, "failed to marshal metadata", zap.Error(err))
			return err
		}
	}

	err = s.db.SaveAuditLog(ctx, models.SaveAuditLogParams{
		UserID: event.UserID,
		AffectedAccountID: uuid.NullUUID{
			UUID:  event.AccountID,
			Valid: event.AccountID != uuid.Nil,
		},
		Action: event.Action.String(),
		Metadata: pqtype.NullRawMessage{
			RawMessage: metadata,
			Valid:      metadata != nil,
		},
	})
	if err != nil {
		logger.Error(ctx, "failed to save audit log", zap.Error(err))
		return err
	}

	logger.Info(ctx, "audit log saved successfully")
	return nil
}

func (s *service) GetAuditLogs(ctx context.Context, accountID uuid.UUID) ([]AuditLog, error) {
	data, err := s.db.GetAuditLogsForAccount(ctx, uuid.NullUUID{
		UUID:  accountID,
		Valid: accountID != uuid.Nil,
	})
	if err != nil {
		return nil, err
	}

	logs := make([]AuditLog, 0, len(data))
	for _, row := range data {
		logs = append(logs, AuditLogFromRow(row))
	}
	return logs, nil
}
