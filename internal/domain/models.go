package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// User represents a system user
type User struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	Email        string     `json:"email" db:"email"`
	PasswordHash string     `json:"-" db:"password_hash"`
	FullName     string     `json:"full_name" db:"full_name"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// Workspace represents a tenant workspace
type Workspace struct {
	ID        uuid.UUID       `json:"id" db:"id"`
	Name      string          `json:"name" db:"name"`
	Slug      string          `json:"slug" db:"slug"`
	Domain    *string         `json:"domain,omitempty" db:"domain"`
	Settings  json.RawMessage `json:"settings" db:"settings"`
	OwnerID   uuid.UUID       `json:"owner_id" db:"owner_id"`
	CreatedAt time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt time.Time       `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time      `json:"deleted_at,omitempty" db:"deleted_at"`
}

// Space represents a top-level container (e.g., "Engineering Docs", "Product Manual")
type Space struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	WorkspaceID uuid.UUID  `json:"workspace_id" db:"workspace_id"`
	Name        string     `json:"name" db:"name"`
	Slug        string     `json:"slug" db:"slug"`
	Description string     `json:"description" db:"description"`
	IconType    string     `json:"icon_type" db:"icon_type"`
	IconValue   string     `json:"icon_value" db:"icon_value"`
	Visibility  string     `json:"visibility" db:"visibility"` // public, private, workspace
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// Collection represents a folder/category in the hierarchy
type Collection struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	WorkspaceID uuid.UUID  `json:"workspace_id" db:"workspace_id"`
	SpaceID     uuid.UUID  `json:"space_id" db:"space_id"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty" db:"parent_id"`
	Name        string     `json:"name" db:"name"`
	Slug        string     `json:"slug" db:"slug"`
	Path        string     `json:"path" db:"path"` // Materialized Path
	Position    int        `json:"position" db:"position"`
	IconType    string     `json:"icon_type" db:"icon_type"`
	IconValue   string     `json:"icon_value" db:"icon_value"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// Page represents a document
type Page struct {
	ID                 uuid.UUID  `json:"id" db:"id"`
	WorkspaceID        uuid.UUID  `json:"workspace_id" db:"workspace_id"`
	SpaceID            uuid.UUID  `json:"space_id" db:"space_id"`
	ParentCollectionID *uuid.UUID `json:"parent_collection_id,omitempty" db:"parent_collection_id"`
	ParentPageID       *uuid.UUID `json:"parent_page_id,omitempty" db:"parent_page_id"`
	Title              string     `json:"title" db:"title"`
	Slug               string     `json:"slug" db:"slug"`
	Path               string     `json:"path" db:"path"` // Materialized Path
	Position           int        `json:"position" db:"position"`
	IconType           string     `json:"icon_type" db:"icon_type"`
	IconValue          string     `json:"icon_value" db:"icon_value"`
	IsPublished        bool       `json:"is_published" db:"is_published"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt          *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// PageVersion represents an immutable version of a page's content
type PageVersion struct {
	ID            uuid.UUID       `json:"id" db:"id"`
	PageID        uuid.UUID       `json:"page_id" db:"page_id"`
	VersionNumber int             `json:"version_number" db:"version_number"`
	Content       json.RawMessage `json:"content" db:"content"` // Structured JSON
	Checksum      string          `json:"checksum" db:"checksum"`
	CreatedBy     uuid.UUID       `json:"created_by" db:"created_by"`
	CreatedAt     time.Time       `json:"created_at" db:"created_at"`
}

// Deployment represents a static site build/publish event
type Deployment struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	WorkspaceID   uuid.UUID  `json:"workspace_id" db:"workspace_id"`
	SiteID        uuid.UUID  `json:"site_id" db:"site_id"`
	EnvironmentID uuid.UUID  `json:"environment_id" db:"environment_id"`
	Status        string     `json:"status" db:"status"` // pending, building, success, failed
	CommitHash    string     `json:"commit_hash" db:"commit_hash"`
	StoragePath   string     `json:"storage_path" db:"storage_path"`
	URL           string     `json:"url" db:"url"`
	Logs          string     `json:"logs" db:"logs"`
	TriggeredBy   uuid.UUID  `json:"triggered_by" db:"triggered_by"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	FinishedAt    *time.Time `json:"finished_at,omitempty" db:"finished_at"`
}

// AuditLog represents a system action record
type AuditLog struct {
	ID           uuid.UUID       `json:"id" db:"id"`
	WorkspaceID  uuid.UUID       `json:"workspace_id" db:"workspace_id"`
	UserID       uuid.UUID       `json:"user_id" db:"user_id"`
	Action       string          `json:"action" db:"action"`
	MetadataJSON json.RawMessage `json:"metadata_json" db:"metadata_json"`
	CreatedAt    time.Time       `json:"created_at" db:"created_at"`
}

// Permission represents an ACL entry
type Permission struct {
	ID           uuid.UUID `json:"id" db:"id"`
	WorkspaceID  uuid.UUID `json:"workspace_id" db:"workspace_id"`
	SubjectType  string    `json:"subject_type" db:"subject_type"` // user, role
	SubjectID    uuid.UUID `json:"subject_id" db:"subject_id"`
	ResourceType string    `json:"resource_type" db:"resource_type"` // workspace, collection, page
	ResourceID   uuid.UUID `json:"resource_id" db:"resource_id"`
	Action       string    `json:"action" db:"action"` // read, write, publish, delete
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// Site represents a documentation site (GitBook-like top level)
type Site struct {
	ID                 uuid.UUID  `json:"id" db:"id"`
	WorkspaceID        uuid.UUID  `json:"workspace_id" db:"workspace_id"`
	Name               string     `json:"name" db:"name"`
	Slug               string     `json:"slug" db:"slug"`
	Plan               string     `json:"plan" db:"plan"`
	DefaultEnvironment string     `json:"default_environment" db:"default_environment"`
	IsPublic           bool       `json:"is_public" db:"is_public"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt          *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// Domain represents a custom domain
type Domain struct {
	ID                uuid.UUID `json:"id" db:"id"`
	WorkspaceID       uuid.UUID `json:"workspace_id" db:"workspace_id"`
	SiteID            uuid.UUID `json:"site_id" db:"site_id"`
	Domain            string    `json:"domain" db:"domain"`
	Type              string    `json:"type" db:"type"` // custom, subdirectory
	Status            string    `json:"status" db:"status"`
	DNSVerified       bool      `json:"dns_verified" db:"dns_verified"`
	SSLStatus         string    `json:"ssl_status" db:"ssl_status"`
	VerificationToken string    `json:"verification_token" db:"verification_token"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
}

// Blob represents content storage (Git engine)
type Blob struct {
	Hash        string          `json:"hash" db:"hash"`
	ContentJSON json.RawMessage `json:"content_json" db:"content_json"`
	SizeBytes   int64           `json:"size_bytes" db:"size_bytes"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
}

// Commit represents a version snapshot (Git engine)
type Commit struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	WorkspaceID uuid.UUID  `json:"workspace_id" db:"workspace_id"`
	SiteID      *uuid.UUID `json:"site_id,omitempty" db:"site_id"`
	TreeHash    string     `json:"tree_hash" db:"tree_hash"`
	ParentHash  *uuid.UUID `json:"parent_hash,omitempty" db:"parent_hash"`
	MergeParentHash *uuid.UUID `json:"merge_parent_hash,omitempty" db:"merge_parent_hash"`
	Message     string     `json:"message" db:"message"`
	AuthorID    uuid.UUID  `json:"author_id" db:"author_id"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
}

// Tree represents a file structure snapshot (Git engine)
type Tree struct {
	ID       uuid.UUID `json:"id" db:"id"`
	CommitID uuid.UUID `json:"commit_id" db:"commit_id"`
	Path     string    `json:"path" db:"path"`
	BlobHash string    `json:"blob_hash" db:"blob_hash"`
	Type     string    `json:"type" db:"type"` // blob, tree
	Mode     string    `json:"mode" db:"mode"`
}

// Branch represents a pointer to a commit (Git engine)
type Branch struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	WorkspaceID  uuid.UUID  `json:"workspace_id" db:"workspace_id"`
	SiteID       uuid.UUID  `json:"site_id" db:"site_id"`
	Name         string     `json:"name" db:"name"`
	HeadCommitID *uuid.UUID `json:"head_commit_id,omitempty" db:"head_commit_id"`
	IsProtected  bool       `json:"is_protected" db:"is_protected"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

// Environment represents a deployment target
type Environment struct {
	ID          uuid.UUID `json:"id" db:"id"`
	WorkspaceID uuid.UUID `json:"workspace_id" db:"workspace_id"`
	SiteID      uuid.UUID `json:"site_id" db:"site_id"`
	Name        string    `json:"name" db:"name"`
	BranchID    uuid.UUID `json:"branch_id" db:"branch_id"`
	URL         string    `json:"url" db:"url"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// Condition represents dynamic content rules
type Condition struct {
	ID          uuid.UUID       `json:"id" db:"id"`
	WorkspaceID uuid.UUID       `json:"workspace_id" db:"workspace_id"`
	Name        string          `json:"name" db:"name"`
	RuleJSON    json.RawMessage `json:"rule_json" db:"rule_json"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
}

// PageTranslation represents multi-language content
type PageTranslation struct {
	ID           uuid.UUID       `json:"id" db:"id"`
	PageID       uuid.UUID       `json:"page_id" db:"page_id"`
	LanguageCode string          `json:"language_code" db:"language_code"`
	Content      json.RawMessage `json:"content" db:"content_json"`
	CreatedAt    time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at" db:"updated_at"`
}

// Plugin represents an extension
type Plugin struct {
	ID          uuid.UUID       `json:"id" db:"id"`
	WorkspaceID uuid.UUID       `json:"workspace_id" db:"workspace_id"`
	Name        string          `json:"name" db:"name"`
	Type        string          `json:"type" db:"type"`
	Config      json.RawMessage `json:"config" db:"config_json"`
	IsEnabled   bool            `json:"is_enabled" db:"is_enabled"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
}

// Webhook represents an event subscription
type Webhook struct {
	ID          uuid.UUID `json:"id" db:"id"`
	WorkspaceID uuid.UUID `json:"workspace_id" db:"workspace_id"`
	Event       string    `json:"event" db:"event"`
	URL         string    `json:"url" db:"url"`
	Secret      string    `json:"secret" db:"secret"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// DocumentContent represents the structured JSON schema (ProseMirror-like)
type DocumentContent struct {
	Type    string  `json:"type"`
	Content []Block `json:"content,omitempty"`
}

type Block struct {
	Type    string                 `json:"type"`
	Attrs   map[string]interface{} `json:"attrs,omitempty"`
	Content []Block                `json:"content,omitempty"`
	Text    string                 `json:"text,omitempty"`
}
