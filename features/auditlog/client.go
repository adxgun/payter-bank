//go:generate mockgen -source=client.go -destination=client_mock.go -package=auditlog

package auditlog

import (
	"github.com/hibiken/asynq"
	"payter-bank/internal/config"
)

// Client is an asyncq client that publishes auditlog tasks to be consumed and processed
// by auditLogProcessor.
// this separation keeps everything small and allows easy unit testing.
type Client interface {
	Enqueue(task *asynq.Task, opt ...asynq.Option) (*asynq.TaskInfo, error)
}

func NewClient(cfg config.RedisConfig) Client {
	return asynq.NewClient(asynq.RedisClientOpt{Addr: cfg.Addr})
}
