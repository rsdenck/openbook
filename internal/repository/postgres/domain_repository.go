package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"openbook/internal/domain"
	"openbook/internal/repository"

	"github.com/google/uuid"
)

type DomainRepository struct {
	db *sql.DB
}

func NewDomainRepository(db *sql.DB) repository.DomainRepository {
	return &DomainRepository{db: db}
}

func (r *DomainRepository) Create(ctx context.Context, d *domain.Domain) error {
	query := `
		INSERT INTO domains (id, workspace_id, site_id, domain, type, status, dns_verified, ssl_status, verification_token, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.db.ExecContext(ctx, query,
		d.ID, d.WorkspaceID, d.SiteID, d.Domain, d.Type, d.Status, d.DNSVerified, d.SSLStatus, d.VerificationToken, d.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create domain: %w", err)
	}
	return nil
}

func (r *DomainRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Domain, error) {
	query := `
		SELECT id, workspace_id, site_id, domain, type, status, dns_verified, ssl_status, verification_token, created_at
		FROM domains WHERE id = $1
	`
	d := &domain.Domain{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&d.ID, &d.WorkspaceID, &d.SiteID, &d.Domain, &d.Type, &d.Status, &d.DNSVerified, &d.SSLStatus, &d.VerificationToken, &d.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("domain not found")
		}
		return nil, fmt.Errorf("failed to get domain: %w", err)
	}
	return d, nil
}

func (r *DomainRepository) ListBySite(ctx context.Context, siteID uuid.UUID) ([]domain.Domain, error) {
	query := `
		SELECT id, workspace_id, site_id, domain, type, status, dns_verified, ssl_status, verification_token, created_at
		FROM domains WHERE site_id = $1
	`
	rows, err := r.db.QueryContext(ctx, query, siteID)
	if err != nil {
		return nil, fmt.Errorf("failed to list domains: %w", err)
	}
	defer rows.Close()

	var domains []domain.Domain
	for rows.Next() {
		var d domain.Domain
		if err := rows.Scan(&d.ID, &d.WorkspaceID, &d.SiteID, &d.Domain, &d.Type, &d.Status, &d.DNSVerified, &d.SSLStatus, &d.VerificationToken, &d.CreatedAt); err != nil {
			return nil, err
		}
		domains = append(domains, d)
	}
	return domains, nil
}
