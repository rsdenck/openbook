package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"openbook/internal/domain"
	"openbook/internal/repository"

	"github.com/google/uuid"
)

type GitRepository struct {
	db *sql.DB
}

func NewGitRepository(db *sql.DB) repository.GitRepository {
	return &GitRepository{db: db}
}

func (r *GitRepository) CreateBlob(ctx context.Context, blob *domain.Blob) error {
	query := `
		INSERT INTO blobs (hash, content_json, size_bytes, created_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (hash) DO NOTHING
	`
	_, err := r.db.ExecContext(ctx, query, blob.Hash, blob.ContentJSON, blob.SizeBytes, blob.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create blob: %w", err)
	}
	return nil
}

func (r *GitRepository) GetBlob(ctx context.Context, hash string) (*domain.Blob, error) {
	query := `SELECT hash, content_json, size_bytes, created_at FROM blobs WHERE hash = $1`
	b := &domain.Blob{}
	var contentJSON []byte
	err := r.db.QueryRowContext(ctx, query, hash).Scan(&b.Hash, &contentJSON, &b.SizeBytes, &b.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("blob not found")
		}
		return nil, fmt.Errorf("failed to get blob: %w", err)
	}
	b.ContentJSON = json.RawMessage(contentJSON)
	return b, nil
}

func (r *GitRepository) CreateCommit(ctx context.Context, commit *domain.Commit) error {
	query := `
		INSERT INTO commits (id, workspace_id, site_id, tree_hash, parent_hash, merge_parent_hash, message, author_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.ExecContext(ctx, query,
		commit.ID, commit.WorkspaceID, commit.SiteID, commit.TreeHash, commit.ParentHash, commit.MergeParentHash, commit.Message, commit.AuthorID, commit.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create commit: %w", err)
	}
	return nil
}

func (r *GitRepository) GetCommit(ctx context.Context, id uuid.UUID) (*domain.Commit, error) {
	query := `
		SELECT id, workspace_id, site_id, tree_hash, parent_hash, merge_parent_hash, message, author_id, created_at
		FROM commits WHERE id = $1
	`
	c := &domain.Commit{}
	var siteID uuid.NullUUID
	var parentHash uuid.NullUUID
	var mergeParentHash uuid.NullUUID
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&c.ID, &c.WorkspaceID, &siteID, &c.TreeHash, &parentHash, &mergeParentHash, &c.Message, &c.AuthorID, &c.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("commit not found")
		}
		return nil, fmt.Errorf("failed to get commit: %w", err)
	}
	if siteID.Valid {
		c.SiteID = &siteID.UUID
	}
	if parentHash.Valid {
		c.ParentHash = &parentHash.UUID
	}
	if mergeParentHash.Valid {
		c.MergeParentHash = &mergeParentHash.UUID
	}
	return c, nil
}

func (r *GitRepository) CreateTree(ctx context.Context, tree *domain.Tree) error {
	query := `
		INSERT INTO trees (id, commit_id, path, blob_hash, type, mode)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.ExecContext(ctx, query,
		tree.ID, tree.CommitID, tree.Path, tree.BlobHash, tree.Type, tree.Mode,
	)
	if err != nil {
		return fmt.Errorf("failed to create tree: %w", err)
	}
	return nil
}

func (r *GitRepository) GetTree(ctx context.Context, commitID uuid.UUID) ([]domain.Tree, error) {
	query := `
		SELECT id, commit_id, path, blob_hash, type, mode
		FROM trees WHERE commit_id = $1
	`
	rows, err := r.db.QueryContext(ctx, query, commitID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tree: %w", err)
	}
	defer rows.Close()

	var trees []domain.Tree
	for rows.Next() {
		var t domain.Tree
		if err := rows.Scan(&t.ID, &t.CommitID, &t.Path, &t.BlobHash, &t.Type, &t.Mode); err != nil {
			return nil, err
		}
		trees = append(trees, t)
	}
	return trees, nil
}

func (r *GitRepository) CreateBranch(ctx context.Context, branch *domain.Branch) error {
	query := `
		INSERT INTO branches (id, workspace_id, site_id, name, head_commit_id, is_protected, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.ExecContext(ctx, query,
		branch.ID, branch.WorkspaceID, branch.SiteID, branch.Name, branch.HeadCommitID, branch.IsProtected, branch.CreatedAt, branch.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create branch: %w", err)
	}
	return nil
}

func (r *GitRepository) GetBranch(ctx context.Context, siteID uuid.UUID, name string) (*domain.Branch, error) {
	query := `
		SELECT id, workspace_id, site_id, name, head_commit_id, is_protected, created_at, updated_at
		FROM branches WHERE site_id = $1 AND name = $2
	`
	b := &domain.Branch{}
	var headCommitID uuid.NullUUID
	err := r.db.QueryRowContext(ctx, query, siteID, name).Scan(
		&b.ID, &b.WorkspaceID, &b.SiteID, &b.Name, &headCommitID, &b.IsProtected, &b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("branch not found")
		}
		return nil, fmt.Errorf("failed to get branch: %w", err)
	}
	if headCommitID.Valid {
		b.HeadCommitID = &headCommitID.UUID
	}
	return b, nil
}

func (r *GitRepository) UpdateBranch(ctx context.Context, branch *domain.Branch) error {
	query := `
		UPDATE branches SET head_commit_id = $1, updated_at = $2 WHERE id = $3
	`
	_, err := r.db.ExecContext(ctx, query, branch.HeadCommitID, branch.UpdatedAt, branch.ID)
	if err != nil {
		return fmt.Errorf("failed to update branch: %w", err)
	}
	return nil
}
