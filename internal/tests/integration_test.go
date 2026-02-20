package tests

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"openbook/internal/domain"
	"openbook/internal/repository/postgres"
	"openbook/internal/service"
	"openbook/internal/usecase"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// IntegrationTestSuite runs a full flow test against real DB and Redis
func TestIntegration_FullFlow(t *testing.T) {
	// 1. Setup Environment
	dbDSN := os.Getenv("TEST_DB_DSN")
	redisAddr := os.Getenv("TEST_REDIS_ADDR")

	if dbDSN == "" || redisAddr == "" {
		t.Skip("Skipping integration test: TEST_DB_DSN or TEST_REDIS_ADDR not set")
	}

	// 2. Connect to Postgres
	db, err := sql.Open("postgres", dbDSN)
	require.NoError(t, err)
	defer db.Close()

	// Verify connection
	err = db.Ping()
	require.NoError(t, err)

	// Clean up tables before test (Dangerous! Only for test DB)
	_, err = db.Exec(`
		TRUNCATE TABLE audit_logs, deployments, commits, trees, blobs, branches, sites, workspaces, users, environments CASCADE
	`)
	require.NoError(t, err)

	// 3. Connect to Redis
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	defer rdb.Close()

	ctx := context.Background()
	err = rdb.Ping(ctx).Err()
	require.NoError(t, err)

	// 4. Setup Dependencies
	auditRepo := postgres.NewAuditLogRepository(db)
	deployRepo := postgres.NewDeploymentRepository(db)
	gitRepo := postgres.NewGitRepository(db)
	publisher := service.NewPublisher(rdb)

	deployUC := usecase.NewDeploymentUseCase(deployRepo, auditRepo, publisher)
	gitUC := usecase.NewGitUseCase(gitRepo)

	// 5. Run Scenario:
	//    Create User -> Create Workspace -> Create Site -> Create Branch -> Commit -> Deploy -> Check Audit -> Check Redis

	// 5.1 Create User (Manual DB insert)
	ownerID := uuid.New()
	_, err = db.Exec(`
		INSERT INTO users (id, email, password_hash, full_name, created_at, updated_at)
		VALUES ($1, 'test@openbook.dev', 'hash', 'Test User', NOW(), NOW())
	`, ownerID)
	require.NoError(t, err)

	// 5.2 Create Workspace (Manual DB insert as we don't have UC for it yet)
	workspaceID := uuid.New()
	_, err = db.Exec(`
		INSERT INTO workspaces (id, name, slug, settings, owner_id, created_at, updated_at)
		VALUES ($1, 'Test Workspace', 'test-ws', '{}', $2, NOW(), NOW())
	`, workspaceID, ownerID)
	require.NoError(t, err)

	// 5.3 Create Site (Manual DB insert)
	siteID := uuid.New()
	_, err = db.Exec(`
		INSERT INTO sites (id, workspace_id, name, slug, plan, default_environment, is_public, created_at, updated_at)
		VALUES ($1, $2, 'Test Site', 'test-site', 'free', 'production', true, NOW(), NOW())
	`, siteID, workspaceID)
	require.NoError(t, err)

	// 5.3 Create Branch (using GitUC)
	// First commit? No, create branch first requires initial commit or empty?
	// The implementation checks "if branch exists". If not, create.
	// But it expects "fromCommitID". If nil, it's an orphan branch.
	branchName := "main"
	branch, err := gitUC.CreateBranch(ctx, workspaceID, siteID, branchName, nil)
	require.NoError(t, err)
	assert.NotNil(t, branch)
	assert.Equal(t, branchName, branch.Name)

	// 5.4 Commit Changes (using GitUC)
	files := map[string][]byte{
		"index.html": []byte("<html>Hello World</html>"),
		"README.md":  []byte("# Readme"),
	}
	commit, err := gitUC.CommitChanges(ctx, workspaceID, siteID, branchName, "Initial commit", ownerID, files)
	require.NoError(t, err)
	assert.NotNil(t, commit)
	assert.NotEmpty(t, commit.TreeHash)

	// 5.4.1 Create Environment (Manual DB insert)
	envID := uuid.New()
	_, err = db.Exec(`
		INSERT INTO environments (id, workspace_id, site_id, name, branch_id, is_active, created_at)
		VALUES ($1, $2, $3, 'production', $4, true, NOW())
	`, envID, workspaceID, siteID, branch.ID)
	require.NoError(t, err)

	// 5.5 Deploy (using DeployUC)
	deployment := &domain.Deployment{
		WorkspaceID:   workspaceID,
		SiteID:        siteID,
		EnvironmentID: envID,
		CommitHash:    commit.ID.String(), // Use Commit ID as hash reference
		TriggeredBy:   ownerID,
	}

	err = deployUC.Create(ctx, deployment)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, deployment.ID)
	assert.Equal(t, "pending", deployment.Status)

	// 5.6 Verify Audit Log
	logs, err := auditRepo.List(ctx, workspaceID, 10, 0)
	require.NoError(t, err)
	assert.NotEmpty(t, logs)
	assert.Equal(t, "deployment.create", logs[0].Action)

	// 5.7 Verify Redis Stream
	// Read from stream to verify message was published
	streams, err := rdb.XRead(ctx, &redis.XReadArgs{
		Streams: []string{"deployments:events", "0"},
		Count:   1,
		Block:   1 * time.Second,
	}).Result()
	require.NoError(t, err)
	assert.NotEmpty(t, streams)

	msg := streams[0].Messages[0]
	var event struct {
		DeploymentID string `json:"deployment_id"`
	}
	// The payload is stored in "payload" field as string
	payloadStr, ok := msg.Values["payload"].(string)
	if !ok {
		// Maybe it's stored directly? The Publisher uses `map[string]interface{}{"payload": jsonBytes}`
		// Let's check publisher implementation.
		// It does: r.client.XAdd(ctx, &redis.XAddArgs{Values: map[string]interface{}{"payload": payload}})
		// So yes, it's in "payload".
	}
	err = json.Unmarshal([]byte(payloadStr), &event)
	require.NoError(t, err)
	assert.Equal(t, deployment.ID.String(), event.DeploymentID)

	fmt.Println("Integration Test Passed Successfully!")
}
