package worker

import (
	"context"
	"log"
	"time"

	"openbook/internal/config"
	"openbook/internal/repository"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Worker struct {
	redisClient *redis.Client
	processor   *DeploymentProcessor
}

func NewWorker(r *redis.Client, dRepo repository.DeploymentRepository, gRepo repository.GitRepository, cfg *config.Config) *Worker {
	processor := NewDeploymentProcessor(dRepo, gRepo, cfg.StoragePath)
	return &Worker{
		redisClient: r,
		processor:   processor,
	}
}

func (w *Worker) Start(ctx context.Context) {
	log.Println("Worker started. Listening to deployments_stream...")

	// Ensure group exists
	err := w.redisClient.XGroupCreateMkStream(ctx, "deployments_stream", "deployment_group", "$").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		log.Printf("Failed to create consumer group: %v", err)
		// Continue anyway? Or fatal?
		// If group fails, we can't consume.
		// But maybe it's a transient error.
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
			entries, err := w.redisClient.XReadGroup(ctx, &redis.XReadGroupArgs{
				Group:    "deployment_group",
				Consumer: "worker-1",
				Streams:  []string{"deployments_stream", ">"},
				Count:    1,
				Block:    0,
			}).Result()

			if err != nil {
				log.Printf("Failed to read from stream: %v", err)
				time.Sleep(1 * time.Second)
				continue
			}

			for _, entry := range entries {
				for _, msg := range entry.Messages {
					deploymentIDStr, ok := msg.Values["deployment_id"].(string)
					if !ok {
						log.Printf("Invalid message format: %v", msg.Values)
						w.ack(ctx, msg.ID)
						continue
					}

					deploymentID, err := uuid.Parse(deploymentIDStr)
					if err != nil {
						log.Printf("Invalid deployment ID: %s", deploymentIDStr)
						w.ack(ctx, msg.ID)
						continue
					}

					if err := w.processor.Process(ctx, deploymentID); err != nil {
						log.Printf("Failed to process deployment %s: %v", deploymentID, err)
						// Don't ACK on failure to allow retry
					} else {
						w.ack(ctx, msg.ID)
					}
				}
			}
		}
	}
}

func (w *Worker) ack(ctx context.Context, msgID string) {
	w.redisClient.XAck(ctx, "deployments_stream", "deployment_group", msgID)
}

func (w *Worker) Stop(ctx context.Context) error {
	log.Println("Worker stopping...")
	return w.redisClient.Close()
}
