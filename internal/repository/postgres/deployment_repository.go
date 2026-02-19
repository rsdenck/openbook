package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"openbook/internal/domain"
	"openbook/internal/repository"

	"github.com/google/uuid"
)

type DeploymentRepository struct {
	db *sql.DB
}

func NewDeploymentRepository(db *sql.DB) repository.DeploymentRepository {
	return &DeploymentRepository{db: db}
}

func (r *DeploymentRepository) Create(ctx context.Context, d *domain.Deployment) error {
	query := `
		INSERT INTO deployments (id, workspace_id, site_id, environment_id, status, commit_hash, triggered_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.ExecContext(ctx, query,
		d.ID, d.WorkspaceID, d.SiteID, d.EnvironmentID, d.Status, d.CommitHash, d.TriggeredBy, d.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create deployment: %w", err)
	}
	return nil
}

func (r *DeploymentRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `UPDATE deployments SET status = $1, finished_at = $2 WHERE id = $3`
	var finishedAt *time.Time
	if status == "success" || status == "failed" {
		now := time.Now()
		finishedAt = &now
	}
	_, err := r.db.ExecContext(ctx, query, status, finishedAt, id)
	if err != nil {
		return fmt.Errorf("failed to update deployment status: %w", err)
	}
	return nil
}

func (r *DeploymentRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Deployment, error) {
	query := `
		SELECT id, workspace_id, site_id, environment_id, status, commit_hash, storage_path, url, logs, triggered_by, created_at, finished_at
		FROM deployments WHERE id = $1
	`
	d := &domain.Deployment{}
	var storagePath, url, logs sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&d.ID, &d.WorkspaceID, &d.SiteID, &d.EnvironmentID, &d.Status, &d.CommitHash, &storagePath, &url, &logs, &d.TriggeredBy, &d.CreatedAt, &d.FinishedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("deployment not found")
		}
		return nil, fmt.Errorf("failed to get deployment: %w", err)
	}
	d.StoragePath = storagePath.String
	d.URL = url.String
	d.Logs = logs.String
	return d, nil
}
