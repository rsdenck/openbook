package repository

import (
	"context"

	"openbook/internal/domain"

	"github.com/google/uuid"
)

type PageRepository interface {
	Create(ctx context.Context, page *domain.Page) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Page, error)
	Update(ctx context.Context, page *domain.Page) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListBySpace(ctx context.Context, spaceID uuid.UUID) ([]domain.Page, error)
	CreateVersion(ctx context.Context, version *domain.PageVersion) error
	GetLatestVersion(ctx context.Context, pageID uuid.UUID) (*domain.PageVersion, error)
}

type CollectionRepository interface {
	Create(ctx context.Context, collection *domain.Collection) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Collection, error)
	ListBySpace(ctx context.Context, spaceID uuid.UUID) ([]domain.Collection, error)
}

type DeploymentRepository interface {
	Create(ctx context.Context, deployment *domain.Deployment) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Deployment, error)
}

type SpaceRepository interface {
	Create(ctx context.Context, space *domain.Space) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Space, error)
}

// New Ultimate Interfaces

type SiteRepository interface {
	Create(ctx context.Context, site *domain.Site) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Site, error)
	GetBySlug(ctx context.Context, workspaceID uuid.UUID, slug string) (*domain.Site, error)
	List(ctx context.Context, workspaceID uuid.UUID) ([]domain.Site, error)
}

type DomainRepository interface {
	Create(ctx context.Context, domain *domain.Domain) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Domain, error)
	ListBySite(ctx context.Context, siteID uuid.UUID) ([]domain.Domain, error)
}

type GitRepository interface {
	CreateBlob(ctx context.Context, blob *domain.Blob) error
	GetBlob(ctx context.Context, hash string) (*domain.Blob, error)
	CreateCommit(ctx context.Context, commit *domain.Commit) error
	GetCommit(ctx context.Context, id uuid.UUID) (*domain.Commit, error)
	CreateTree(ctx context.Context, tree *domain.Tree) error
	GetTree(ctx context.Context, commitID uuid.UUID) ([]domain.Tree, error)
	CreateBranch(ctx context.Context, branch *domain.Branch) error
	GetBranch(ctx context.Context, siteID uuid.UUID, name string) (*domain.Branch, error)
	UpdateBranch(ctx context.Context, branch *domain.Branch) error
}

type AuditLogRepository interface {
	Create(ctx context.Context, log *domain.AuditLog) error
	List(ctx context.Context, workspaceID uuid.UUID, limit, offset int) ([]domain.AuditLog, error)
}
