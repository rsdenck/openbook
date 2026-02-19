-- Enable pg_trgm for text search if needed later
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- 1. Sites Manager (Replaces strict Space hierarchy or sits above it)
CREATE TABLE sites (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL,
    plan VARCHAR(50) DEFAULT 'basic', -- basic, premium, ultimate
    default_environment VARCHAR(50) DEFAULT 'production',
    is_public BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(workspace_id, slug)
);

-- 2. Domain Manager
CREATE TABLE domains (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id),
    site_id UUID NOT NULL REFERENCES sites(id),
    domain VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL CHECK (type IN ('custom', 'subdirectory')),
    status VARCHAR(50) DEFAULT 'pending', -- pending, active, error
    dns_verified BOOLEAN DEFAULT FALSE,
    ssl_status VARCHAR(50) DEFAULT 'none', -- none, provisioning, active, error
    verification_token VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(domain)
);

-- 3. Git Engine: Blobs (Content storage, deduplicated by hash)
CREATE TABLE blobs (
    hash VARCHAR(64) PRIMARY KEY, -- SHA256 of content
    content_json JSONB NOT NULL,
    size_bytes BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 4. Git Engine: Commits
CREATE TABLE commits (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id),
    site_id UUID REFERENCES sites(id), -- Optional link to site
    tree_hash VARCHAR(64) NOT NULL, -- Merkle tree root hash
    parent_hash UUID REFERENCES commits(id),
    merge_parent_hash UUID REFERENCES commits(id), -- For merge commits
    message TEXT NOT NULL,
    author_id UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_commits_workspace ON commits(workspace_id);
CREATE INDEX idx_commits_parent ON commits(parent_hash);

-- 5. Git Engine: Trees (Snapshot of structure)
-- Note: In a real Git, trees are recursive. Here we simplify for "File List" per commit
-- or we can model it as: id, commit_id (nullable if shared), path, blob_hash
CREATE TABLE trees (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    commit_id UUID NOT NULL REFERENCES commits(id),
    path TEXT NOT NULL, -- logical path /collection/page
    blob_hash VARCHAR(64) NOT NULL REFERENCES blobs(hash),
    type VARCHAR(20) DEFAULT 'blob', -- blob or tree (for future recursion)
    mode VARCHAR(10) DEFAULT '100644',
    UNIQUE(commit_id, path)
);

-- 6. Branching System
CREATE TABLE branches (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id),
    site_id UUID NOT NULL REFERENCES sites(id),
    name VARCHAR(255) NOT NULL, -- main, staging, feature/xyz
    head_commit_id UUID REFERENCES commits(id),
    is_protected BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(site_id, name)
);

-- 7. Environments
CREATE TABLE environments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id),
    site_id UUID NOT NULL REFERENCES sites(id),
    name VARCHAR(50) NOT NULL, -- production, staging, dev
    branch_id UUID NOT NULL REFERENCES branches(id),
    url VARCHAR(255),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(site_id, name)
);

-- 8. Dynamic Conditions (Personalization)
CREATE TABLE conditions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id),
    name VARCHAR(255) NOT NULL,
    rule_json JSONB NOT NULL, -- { "if": { "user.plan": "enterprise" }, ... }
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 9. Multi-language Support
CREATE TABLE page_translations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    page_id UUID NOT NULL REFERENCES pages(id),
    language_code VARCHAR(10) NOT NULL, -- en, pt-BR, es
    content_json JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(page_id, language_code)
);

-- 10. Plugin System
CREATE TABLE plugins (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id),
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL, -- editor, build, webhook, analytics
    config_json JSONB DEFAULT '{}',
    is_enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 11. Webhooks
CREATE TABLE webhooks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id),
    event VARCHAR(255) NOT NULL, -- page.updated, commit.created, etc.
    url TEXT NOT NULL,
    secret VARCHAR(255),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 12. Enterprise Security: Audit Logs
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id),
    user_id UUID NOT NULL REFERENCES users(id),
    action VARCHAR(255) NOT NULL,
    metadata_json JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 13. Update Deployments Table to match new architecture
ALTER TABLE deployments ADD COLUMN IF NOT EXISTS site_id UUID REFERENCES sites(id);
ALTER TABLE deployments ADD COLUMN IF NOT EXISTS environment_id UUID REFERENCES environments(id);
ALTER TABLE deployments ADD COLUMN IF NOT EXISTS commit_hash VARCHAR(64);
-- Make existing columns nullable if needed or keep them for backward compatibility

-- Indexes for performance
CREATE INDEX idx_sites_workspace ON sites(workspace_id);
CREATE INDEX idx_domains_workspace ON domains(workspace_id);
CREATE INDEX idx_branches_site ON branches(site_id);
CREATE INDEX idx_trees_commit ON trees(commit_id);
CREATE INDEX idx_audit_logs_workspace ON audit_logs(workspace_id);

-- Optimistic Locking helper for Pages (if not already handled)
ALTER TABLE pages ADD COLUMN IF NOT EXISTS version_lock INT DEFAULT 1;
