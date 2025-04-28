package auditlog

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/sqlc-dev/pqtype"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"payter-bank/internal/config"
	"payter-bank/internal/database/models"
	databasemocks "payter-bank/internal/database/models/mocks"
	"testing"
)

func TestService_Submit(t *testing.T) {
	t.Run("should submit event to audit log", func(t *testing.T) {
		userID := uuid.MustParse("12345678-1234-1234-1234-123456789012")
		accountID := uuid.MustParse("12345678-1234-1234-1234-123456789012")
		ev := Event{
			Action:    "test_action",
			UserID:    userID,
			AccountID: accountID,
			Metadata:  map[string]interface{}{"test_key": "test_value"},
		}
		expectedPayload, _ := json.Marshal(ev)
		expectedTask := asynq.NewTask(auditLogTaskName, expectedPayload)

		mocker := newAuditLogServiceMocker(t)
		mocker.client.EXPECT().Enqueue(expectedTask).
			Return(&asynq.TaskInfo{}, nil)

		err := mocker.service.Submit(context.TODO(), ev)
		assert.NoError(t, err)
	})

	t.Run("should fail when audit log save fails", func(t *testing.T) {
		userID := uuid.MustParse("12345678-1234-1234-1234-123456789012")
		accountID := uuid.MustParse("12345678-1234-1234-1234-123456789012")
		ev := Event{
			Action:    "test_action",
			UserID:    userID,
			AccountID: accountID,
			Metadata:  map[string]interface{}{"test_key": "test_value"},
		}
		expectedPayload, _ := json.Marshal(ev)
		expectedTask := asynq.NewTask(auditLogTaskName, expectedPayload)

		mocker := newAuditLogServiceMocker(t)
		mocker.client.EXPECT().Enqueue(expectedTask).
			Return(nil, errors.New("failed to enqueue task"))

		err := mocker.service.Submit(context.TODO(), ev)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to enqueue task")
	})
}

func TestService_ProcessTask(t *testing.T) {
	t.Run("should process task successfully", func(t *testing.T) {
		userID := uuid.MustParse("12345678-1234-1234-1234-123456789012")
		accountID := uuid.MustParse("12345678-1234-1234-1234-123456789012")
		ev := Event{
			Action:    "test_action",
			UserID:    userID,
			AccountID: accountID,
			Metadata:  map[string]interface{}{"test_key": "test_value"},
		}
		taskPayload, _ := json.Marshal(ev)
		metadata, _ := json.Marshal(ev.Metadata)
		task := asynq.NewTask(auditLogTaskName, taskPayload)
		expectedAuditLogSaveParam := models.SaveAuditLogParams{
			UserID: userID,
			AffectedAccountID: uuid.NullUUID{
				UUID:  accountID,
				Valid: true,
			},
			Action: "test_action",
			Metadata: pqtype.NullRawMessage{
				RawMessage: metadata,
				Valid:      true,
			},
		}

		mocker := newAuditLogServiceMocker(t)
		mocker.db.EXPECT().SaveAuditLog(gomock.Any(), expectedAuditLogSaveParam).
			Return(nil)

		err := mocker.processor.ProcessTask(context.TODO(), task)
		assert.NoError(t, err)
	})
	
	t.Run("should process task with nil metadata successfully", func(t *testing.T) {
		userID := uuid.MustParse("12345678-1234-1234-1234-123456789012")
		accountID := uuid.MustParse("12345678-1234-1234-1234-123456789012")
		ev := Event{
			Action:    "test_action",
			UserID:    userID,
			AccountID: accountID,
			Metadata:  nil,
		}
		taskPayload, _ := json.Marshal(ev)
		task := asynq.NewTask(auditLogTaskName, taskPayload)
		expectedAuditLogSaveParam := models.SaveAuditLogParams{
			UserID: userID,
			AffectedAccountID: uuid.NullUUID{
				UUID:  accountID,
				Valid: true,
			},
			Action: "test_action",
			Metadata: pqtype.NullRawMessage{
				RawMessage: nil,
				Valid:      false,
			},
		}
		mocker := newAuditLogServiceMocker(t)
		mocker.db.EXPECT().SaveAuditLog(gomock.Any(), expectedAuditLogSaveParam).
			Return(nil)

		err := mocker.processor.ProcessTask(context.TODO(), task)
		assert.NoError(t, err)
	})
}

type auditLogServiceMocker struct {
	db        *databasemocks.MockQuerier
	client    *MockClient
	cfg       config.Config
	service   Service
	processor auditLogProcessor
}

func newAuditLogServiceMocker(t *testing.T) *auditLogServiceMocker {
	ctrl := gomock.NewController(t)
	db := databasemocks.NewMockQuerier(ctrl)
	client := NewMockClient(ctrl)
	cfg := config.Config{}

	svc := NewService(cfg, client, db)
	return &auditLogServiceMocker{
		db: db, client: client, cfg: cfg, service: svc, processor: newProcessor(db)}
}
