package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"openbook/internal/domain"
	"openbook/internal/repository"

	"github.com/google/uuid"
)

type AuditLogRepository struct {
	db *sql.DB
}

func NewAuditLogRepository(db *sql.DB) repository.AuditLogRepository {
	return &AuditLogRepository{db: db}
}

func (r *AuditLogRepository) Create(ctx context.Context, log *domain.AuditLog) error {
	query := `
		INSERT INTO audit_logs (id, workspace_id, user_id, action, metadata_json, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.ExecContext(ctx, query,
		log.ID, log.WorkspaceID, log.UserID, log.Action, log.MetadataJSON, log.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}
	return nil
}

func (r *AuditLogRepository) List(ctx context.Context, workspaceID uuid.UUID, limit, offset int) ([]domain.AuditLog, error) {
	query := `
		SELECT id, workspace_id, user_id, action, metadata_json, created_at
		FROM audit_logs
		WHERE workspace_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, workspaceID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list audit logs: %w", err)
	}
	defer rows.Close()

	var logs []domain.AuditLog
	for rows.Next() {
		var l domain.AuditLog
		if err := rows.Scan(&l.ID, &l.WorkspaceID, &l.UserID, &l.Action, &l.MetadataJSON, &l.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, nil
}
