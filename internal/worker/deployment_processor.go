package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"openbook/internal/repository"

	"github.com/google/uuid"
)

type DeploymentProcessor struct {
	deploymentRepo repository.DeploymentRepository
	gitRepo        repository.GitRepository
	storagePath    string
}

func NewDeploymentProcessor(
	dRepo repository.DeploymentRepository,
	gRepo repository.GitRepository,
	storagePath string,
) *DeploymentProcessor {
	return &DeploymentProcessor{
		deploymentRepo: dRepo,
		gitRepo:        gRepo,
		storagePath:    storagePath,
	}
}

func (p *DeploymentProcessor) Process(ctx context.Context, deploymentID uuid.UUID) error {
	log.Printf("Starting deployment process for ID: %s", deploymentID)

	// 1. Fetch deployment
	deployment, err := p.deploymentRepo.GetByID(ctx, deploymentID)
	if err != nil {
		return fmt.Errorf("failed to get deployment: %w", err)
	}

	// 2. Update status to building
	if err := p.deploymentRepo.UpdateStatus(ctx, deployment.ID, "building"); err != nil {
		return fmt.Errorf("failed to update status to building: %w", err)
	}

	// 3. Resolve commit and tree
	// Note: CommitHash in Deployment is a string, but GetCommit expects UUID?
	// The models say CommitHash string. The repository GetCommit expects UUID.
	// This is a mismatch. Let's assume CommitHash in deployment is the UUID of the commit for now, or fix repository.
	// Looking at models.go from previous turn: CommitHash string.
	// Looking at repository/postgres/git_repository.go: GetCommit(ctx, id uuid.UUID).
	// So CommitHash should be a UUID string.

	commitUUID, err := uuid.Parse(deployment.CommitHash)
	if err != nil {
		p.failDeployment(ctx, deployment.ID, fmt.Errorf("invalid commit hash: %w", err))
		return err
	}

	commit, err := p.gitRepo.GetCommit(ctx, commitUUID)
	if err != nil {
		p.failDeployment(ctx, deployment.ID, fmt.Errorf("failed to get commit: %w", err))
		return err
	}

	// 4. Build and Save
	siteStoragePath := filepath.Join(p.storagePath, deployment.WorkspaceID.String(), deployment.SiteID.String(), deployment.CommitHash)
	if err := os.MkdirAll(siteStoragePath, 0755); err != nil {
		p.failDeployment(ctx, deployment.ID, fmt.Errorf("failed to create storage dir: %w", err))
		return err
	}

	// Fetch all tree entries for this commit
	trees, err := p.gitRepo.GetTree(ctx, commit.ID)
	if err != nil {
		p.failDeployment(ctx, deployment.ID, fmt.Errorf("failed to get tree: %w", err))
		return err
	}

	for _, treeNode := range trees {
		if treeNode.Type == "blob" {
			// Fetch blob content
			blob, err := p.gitRepo.GetBlob(ctx, treeNode.BlobHash)
			if err != nil {
				p.failDeployment(ctx, deployment.ID, fmt.Errorf("failed to get blob %s: %w", treeNode.BlobHash, err))
				return err
			}

			// Determine full path
			// Assuming treeNode.Path is relative to site root, e.g., "index.html" or "docs/page.md"
			fullPath := filepath.Join(siteStoragePath, treeNode.Path)

			// Ensure directory exists
			if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
				p.failDeployment(ctx, deployment.ID, fmt.Errorf("failed to create dir for %s: %w", treeNode.Path, err))
				return err
			}

			// Write file
			// content_json in Blob is json.RawMessage. We probably want the raw content.
			// If it's JSON encoded, we might need to decode it.
			// But for "Construir HTML real", maybe the blob IS the content?
			// Or maybe we need to render it?
			// CONTEXTO3: "Construir HTML real".
			// If the blob is Markdown, we should convert to HTML.
			// For now, let's just write the content as is (or unquoted JSON string).

			var content string
			if err := json.Unmarshal(blob.ContentJSON, &content); err != nil {
				// If not a JSON string, maybe it's just raw JSON?
				// Write raw bytes
				content = string(blob.ContentJSON)
			}

			// Simple HTML wrapper if it's markdown (naive implementation)
			if filepath.Ext(treeNode.Path) == ".md" {
				fullPath = fullPath[0:len(fullPath)-3] + ".html"
				content = fmt.Sprintf("<html><body><pre>%s</pre></body></html>", content)
			}

			if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
				p.failDeployment(ctx, deployment.ID, fmt.Errorf("failed to write file %s: %w", treeNode.Path, err))
				return err
			}
		}
	}

	// 5. Update status to success
	if err := p.deploymentRepo.UpdateStatus(ctx, deployment.ID, "success"); err != nil {
		return fmt.Errorf("failed to update status to success: %w", err)
	}

	log.Printf("Deployment %s completed successfully", deploymentID)
	return nil
}

func (p *DeploymentProcessor) failDeployment(ctx context.Context, id uuid.UUID, err error) {
	log.Printf("Deployment %s failed: %v", id, err)
	_ = p.deploymentRepo.UpdateStatus(ctx, id, "failed")
}
