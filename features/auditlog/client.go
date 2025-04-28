//go:generate mockgen -source=client.go -destination=client_mock.go -package=auditlog

package auditlog

import (
	"github.com/hibiken/asynq"
	"payter-bank/internal/config"
)

type Client interface {
	Enqueue(task *asynq.Task, opt...asynq.Option) (*asynq.TaskInfo, error)
}

func NewClient(cfg config.RedisConfig) Client {
	return asynq.NewClient(asynq.RedisClientOpt{Addr: cfg.Addr})
}