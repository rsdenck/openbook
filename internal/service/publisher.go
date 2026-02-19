package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Publisher struct {
	redis *redis.Client
}

func NewPublisher(r *redis.Client) *Publisher {
	return &Publisher{redis: r}
}

func (p *Publisher) PublishDeployment(ctx context.Context, deploymentID uuid.UUID) error {
	err := p.redis.XAdd(ctx, &redis.XAddArgs{
		Stream: "deployments_stream",
		Values: map[string]interface{}{
			"deployment_id": deploymentID.String(),
		},
	}).Err()
	if err != nil {
		return fmt.Errorf("failed to publish deployment event: %w", err)
	}
	return nil
}
