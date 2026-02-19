package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"openbook/internal/domain"
	"openbook/internal/repository"
	"openbook/internal/service"

	"github.com/google/uuid"
)

type DeploymentUseCase struct {
	repo      repository.DeploymentRepository
	auditRepo repository.AuditLogRepository
	publisher *service.Publisher
}

func NewDeploymentUseCase(repo repository.DeploymentRepository, auditRepo repository.AuditLogRepository, publisher *service.Publisher) *DeploymentUseCase {
	return &DeploymentUseCase{
		repo:      repo,
		auditRepo: auditRepo,
		publisher: publisher,
	}
}

func (uc *DeploymentUseCase) Create(ctx context.Context, d *domain.Deployment) error {
	d.ID = uuid.New()
	d.Status = "pending"
	d.CreatedAt = time.Now()

	if err := uc.repo.Create(ctx, d); err != nil {
		return fmt.Errorf("failed to create deployment: %w", err)
	}

	// Audit Log
	metadata, _ := json.Marshal(map[string]interface{}{
		"site_id":        d.SiteID,
		"environment_id": d.EnvironmentID,
		"commit_hash":    d.CommitHash,
	})

	audit := &domain.AuditLog{
		ID:           uuid.New(),
		WorkspaceID:  d.WorkspaceID,
		UserID:       d.TriggeredBy,
		Action:       "deployment.create",
		MetadataJSON: metadata,
		CreatedAt:    time.Now(),
	}
	// Log error but don't fail the operation
	if err := uc.auditRepo.Create(ctx, audit); err != nil {
		fmt.Printf("failed to create audit log: %v\n", err)
	}

	if err := uc.publisher.PublishDeployment(ctx, d.ID); err != nil {
		return fmt.Errorf("failed to publish deployment event: %w", err)
	}

	return nil
}

func (uc *DeploymentUseCase) GetByID(ctx context.Context, id uuid.UUID) (*domain.Deployment, error) {
	return uc.repo.GetByID(ctx, id)
}
