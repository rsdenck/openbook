-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Workspaces
CREATE TABLE workspaces (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL UNIQUE,
    domain VARCHAR(255),
    settings JSONB DEFAULT '{}',
    owner_id UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Workspace Members (Roles)
CREATE TABLE workspace_members (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id),
    user_id UUID NOT NULL REFERENCES users(id),
    role VARCHAR(50) NOT NULL CHECK (role IN ('owner', 'admin', 'editor', 'viewer')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(workspace_id, user_id)
);

-- Spaces (Top level container within workspace)
CREATE TABLE spaces (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL,
    description TEXT,
    icon_type VARCHAR(50),
    icon_value VARCHAR(255),
    visibility VARCHAR(50) DEFAULT 'private', -- public, private, workspace
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(workspace_id, slug)
);

-- Collections (Hierarchical)
CREATE TABLE collections (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id),
    space_id UUID NOT NULL REFERENCES spaces(id),
    parent_id UUID REFERENCES collections(id),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL,
    path TEXT NOT NULL, -- Materialized Path: parent_id/id
    position INT DEFAULT 0,
    icon_type VARCHAR(50),
    icon_value VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_collections_path ON collections(path);
CREATE INDEX idx_collections_workspace ON collections(workspace_id);

-- Pages
CREATE TABLE pages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id),
    space_id UUID NOT NULL REFERENCES spaces(id),
    parent_collection_id UUID REFERENCES collections(id),
    parent_page_id UUID REFERENCES pages(id),
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL,
    path TEXT NOT NULL, -- Materialized Path
    position INT DEFAULT 0,
    icon_type VARCHAR(50),
    icon_value VARCHAR(255),
    is_published BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_pages_path ON pages(path);
CREATE INDEX idx_pages_workspace ON pages(workspace_id);

-- Page Versions (Immutable content)
CREATE TABLE page_versions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    page_id UUID NOT NULL REFERENCES pages(id),
    version_number INT NOT NULL,
    content JSONB NOT NULL, -- Structured JSON
    checksum VARCHAR(64) NOT NULL,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(page_id, version_number)
);

-- Deployments
CREATE TABLE deployments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id),
    space_id UUID NOT NULL REFERENCES spaces(id),
    status VARCHAR(50) NOT NULL CHECK (status IN ('pending', 'building', 'success', 'failed')),
    version_hash VARCHAR(64) NOT NULL,
    storage_path VARCHAR(255),
    url VARCHAR(255),
    logs TEXT,
    triggered_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    finished_at TIMESTAMP WITH TIME ZONE
);

-- Permissions (Granular ACL)
CREATE TABLE permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id),
    subject_type VARCHAR(50) NOT NULL, -- user, role
    subject_id UUID NOT NULL,
    resource_type VARCHAR(50) NOT NULL, -- workspace, collection, page
    resource_id UUID NOT NULL,
    action VARCHAR(50) NOT NULL, -- read, write, publish, delete
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_permissions_resource ON permissions(resource_type, resource_id);
