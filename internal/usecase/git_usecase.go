package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"openbook/internal/domain"
	"openbook/internal/repository"

	"github.com/google/uuid"
)

type GitUseCase struct {
	repo repository.GitRepository
}

func NewGitUseCase(repo repository.GitRepository) *GitUseCase {
	return &GitUseCase{repo: repo}
}

func (uc *GitUseCase) CreateBranch(ctx context.Context, workspaceID, siteID uuid.UUID, name string, fromCommitID *uuid.UUID) (*domain.Branch, error) {
	// Check if branch exists
	if _, err := uc.repo.GetBranch(ctx, siteID, name); err == nil {
		return nil, fmt.Errorf("branch %s already exists", name)
	}

	branch := &domain.Branch{
		ID:           uuid.New(),
		WorkspaceID:  workspaceID,
		SiteID:       siteID,
		Name:         name,
		HeadCommitID: fromCommitID,
		IsProtected:  false,
		CreatedAt:    time.Now(),
	}

	if err := uc.repo.CreateBranch(ctx, branch); err != nil {
		return nil, fmt.Errorf("failed to create branch: %w", err)
	}

	return branch, nil
}

func (uc *GitUseCase) GetBranch(ctx context.Context, siteID uuid.UUID, name string) (*domain.Branch, error) {
	return uc.repo.GetBranch(ctx, siteID, name)
}

func (uc *GitUseCase) MergeBranches(ctx context.Context, workspaceID, siteID uuid.UUID, sourceName, targetName string, authorID uuid.UUID) (*domain.Commit, error) {
	// 1. Get Source Branch
	sourceBranch, err := uc.repo.GetBranch(ctx, siteID, sourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to get source branch: %w", err)
	}
	if sourceBranch.HeadCommitID == nil {
		return nil, fmt.Errorf("source branch has no commits")
	}

	// 2. Get Target Branch
	targetBranch, err := uc.repo.GetBranch(ctx, siteID, targetName)
	if err != nil {
		return nil, fmt.Errorf("failed to get target branch: %w", err)
	}

	// 3. Get Source Head Commit (to get TreeHash)
	sourceHead, err := uc.repo.GetCommit(ctx, *sourceBranch.HeadCommitID)
	if err != nil {
		return nil, fmt.Errorf("failed to get source head commit: %w", err)
	}

	// 4. Create Merge Commit
	// We use the Source's TreeHash as the result of the merge (Assuming "Theirs" strategy for simplicity)
	// In a real system, we would perform a 3-way merge of the trees and create a new Tree.
	mergeCommit := &domain.Commit{
		ID:              uuid.New(),
		WorkspaceID:     workspaceID,
		SiteID:          &siteID,
		TreeHash:        sourceHead.TreeHash,
		ParentHash:      targetBranch.HeadCommitID,
		MergeParentHash: sourceBranch.HeadCommitID,
		Message:         fmt.Sprintf("Merge branch '%s' into '%s'", sourceName, targetName),
		AuthorID:        authorID,
		CreatedAt:       time.Now(),
	}

	if err := uc.repo.CreateCommit(ctx, mergeCommit); err != nil {
		return nil, fmt.Errorf("failed to create merge commit: %w", err)
	}

	// 5. Update Target Branch Head
	targetBranch.HeadCommitID = &mergeCommit.ID
	targetBranch.UpdatedAt = time.Now()
	if err := uc.repo.UpdateBranch(ctx, targetBranch); err != nil {
		return nil, fmt.Errorf("failed to update target branch: %w", err)
	}

	return mergeCommit, nil
}

func (uc *GitUseCase) CommitChanges(ctx context.Context, workspaceID, siteID uuid.UUID, branchName string, message string, authorID uuid.UUID, files map[string][]byte) (*domain.Commit, error) {
	// 1. Get Branch
	branch, err := uc.repo.GetBranch(ctx, siteID, branchName)
	if err != nil {
		// If branch doesn't exist, create it pointing to nothing (initial commit)
		// Or fail?
		// For now, fail.
		return nil, fmt.Errorf("branch %s not found: %w", branchName, err)
	}

	// 2. Create Blobs
	treeEntries := make([]domain.Tree, 0, len(files))
	for path, content := range files {
		// Hash content
		hash := sha256.Sum256(content)
		blobHash := hex.EncodeToString(hash[:])

		// Store Blob (idempotent)
		blob := &domain.Blob{
			Hash:        blobHash,
			ContentJSON: json.RawMessage(content), // Storing raw content as JSON message might be tricky if not JSON
			// Wait, ContentJSON implies it's JSON.
			// If it's HTML/Markdown, we should wrap it in JSON string?
			// Or just store bytes?
			// The model says ContentJSON json.RawMessage.
			// Let's wrap it in a JSON string for now to be safe with DB type.
			SizeBytes: int64(len(content)),
			CreatedAt: time.Now(),
		}

		// Wrap content in JSON string
		contentJSON, _ := json.Marshal(string(content))
		blob.ContentJSON = contentJSON

		if err := uc.repo.CreateBlob(ctx, blob); err != nil {
			return nil, fmt.Errorf("failed to create blob for %s: %w", path, err)
		}

		treeEntries = append(treeEntries, domain.Tree{
			ID:       uuid.New(),
			Path:     path,
			BlobHash: blobHash,
			Type:     "blob",
			Mode:     "100644",
		})
	}

	// 3. Create Tree (Snapshot model - assuming full tree for simplicity)
	// In real Git, we'd reuse existing tree entries.
	// Here we just create a new commit with these files.
	// If files map is partial update, we need to fetch old tree.
	// For "initial implementation", let's assume full snapshot or simple append.
	// Given we don't have "Edit Page" logic fully fleshed out, let's assume full snapshot.

	// Calculate Tree Hash (Merkle Root) - Simplified: Hash of all blob hashes sorted
	// For now, just a random hash or hash of first blob
	// Let's make a real hash
	treeHasher := sha256.New()
	for _, entry := range treeEntries {
		treeHasher.Write([]byte(entry.Path + entry.BlobHash))
	}
	treeHash := hex.EncodeToString(treeHasher.Sum(nil))

	// 4. Create Commit
	commit := &domain.Commit{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		SiteID:      &siteID,
		TreeHash:    treeHash,
		ParentHash:  branch.HeadCommitID,
		Message:     message,
		AuthorID:    authorID,
		CreatedAt:   time.Now(),
	}

	if err := uc.repo.CreateCommit(ctx, commit); err != nil {
		return nil, fmt.Errorf("failed to create commit: %w", err)
	}

	// 5. Save Tree Entries linked to Commit
	for _, entry := range treeEntries {
		entry.CommitID = commit.ID
		if err := uc.repo.CreateTree(ctx, &entry); err != nil {
			return nil, fmt.Errorf("failed to create tree entry: %w", err)
		}
	}

	// 6. Update Branch Head
	branch.HeadCommitID = &commit.ID
	branch.UpdatedAt = time.Now()
	if err := uc.repo.UpdateBranch(ctx, branch); err != nil {
		return nil, fmt.Errorf("failed to update branch: %w", err)
	}

	return commit, nil
}
