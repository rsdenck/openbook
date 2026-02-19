package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"openbook/internal/domain"
	"openbook/internal/repository"

	"github.com/google/uuid"
)

type SiteRepository struct {
	db *sql.DB
}

func NewSiteRepository(db *sql.DB) repository.SiteRepository {
	return &SiteRepository{db: db}
}

func (r *SiteRepository) Create(ctx context.Context, site *domain.Site) error {
	query := `
		INSERT INTO sites (id, workspace_id, name, slug, plan, default_environment, is_public, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.ExecContext(ctx, query,
		site.ID, site.WorkspaceID, site.Name, site.Slug, site.Plan, site.DefaultEnvironment, site.IsPublic, site.CreatedAt, site.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create site: %w", err)
	}
	return nil
}

func (r *SiteRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Site, error) {
	query := `
		SELECT id, workspace_id, name, slug, plan, default_environment, is_public, created_at, updated_at
		FROM sites WHERE id = $1
	`
	s := &domain.Site{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&s.ID, &s.WorkspaceID, &s.Name, &s.Slug, &s.Plan, &s.DefaultEnvironment, &s.IsPublic, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("site not found")
		}
		return nil, fmt.Errorf("failed to get site: %w", err)
	}
	return s, nil
}

func (r *SiteRepository) GetBySlug(ctx context.Context, workspaceID uuid.UUID, slug string) (*domain.Site, error) {
	query := `
		SELECT id, workspace_id, name, slug, plan, default_environment, is_public, created_at, updated_at
		FROM sites WHERE workspace_id = $1 AND slug = $2
	`
	s := &domain.Site{}
	err := r.db.QueryRowContext(ctx, query, workspaceID, slug).Scan(
		&s.ID, &s.WorkspaceID, &s.Name, &s.Slug, &s.Plan, &s.DefaultEnvironment, &s.IsPublic, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("site not found")
		}
		return nil, fmt.Errorf("failed to get site: %w", err)
	}
	return s, nil
}

func (r *SiteRepository) List(ctx context.Context, workspaceID uuid.UUID) ([]domain.Site, error) {
	query := `
		SELECT id, workspace_id, name, slug, plan, default_environment, is_public, created_at, updated_at
		FROM sites WHERE workspace_id = $1
	`
	rows, err := r.db.QueryContext(ctx, query, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to list sites: %w", err)
	}
	defer rows.Close()

	var sites []domain.Site
	for rows.Next() {
		var s domain.Site
		if err := rows.Scan(&s.ID, &s.WorkspaceID, &s.Name, &s.Slug, &s.Plan, &s.DefaultEnvironment, &s.IsPublic, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		sites = append(sites, s)
	}
	return sites, nil
}
