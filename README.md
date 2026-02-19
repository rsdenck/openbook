# OPENBOOK

![Build Status](https://img.shields.io/github/actions/workflow/status/rsdenck/openbook/ci.yml?style=flat-square&logo=github&label=build&color=42b883&labelColor=1a1a1a)
![Go Version](https://img.shields.io/github/go-mod/go-version/rsdenck/openbook?style=flat-square&logo=go&label=go&color=42b883&labelColor=1a1a1a)
![License](https://img.shields.io/github/license/rsdenck/openbook?style=flat-square&label=license&color=42b883&labelColor=1a1a1a)

**KERNEL-LEVEL DOCUMENTATION PLATFORM**

OpenBook is a high-performance, git-native documentation engine built for enterprise environments. It operates directly on the Linux kernel, leveraging real system calls, strict process isolation, and a zero-mock architecture to deliver production-grade reliability.

---

## ■ SYSTEM ARCHITECTURE

The system follows strict **Clean Architecture** principles, enforcing separation of concerns between the core domain, use cases, and infrastructure adapters.

```mermaid
graph TD
    A[API Layer] --> B[Use Cases]
    B --> C[Domain Entities]
    B --> D[Repositories]
    E[Worker Service] --> D
    E --> F[Redis Streams]
    D --> G[PostgreSQL]
    D --> H[Git Engine]
```

### › Core Components

*   **API Service**: RESTful interface for workspace management, branching, and deployment triggers.
*   **Worker Service**: Asynchronous execution engine consuming Redis Streams for heavy-lifting tasks (builds, git operations).
*   **Git Engine**: Custom implementation of Git plumbing commands (commits, trees, blobs) storing state in PostgreSQL and the file system.
*   **Deployment Pipeline**: Real-time static site generation and atomic swappable deployments.

---

## ■ TECHNOLOGY STACK

| Component | Technology | Version | Role |
| :--- | :--- | :--- | :--- |
| **Language** | Go | 1.21+ | Core logic, API, Worker |
| **Database** | PostgreSQL | 15+ | Relational data, Git objects, Audit logs |
| **Broker** | Redis | 7+ | Event streaming, Task queues |
| **OS** | Linux | Kernel 5+ | Host environment, Namespaces, FS |
| **CI/CD** | GitHub Actions | - | Automated testing, build, release |

---

## ■ KEY FEATURES

### › Git-Native Versioning
Every change is a commit. OpenBook implements a Git-compatible object storage layer (CAS) within PostgreSQL, allowing for:
*   Atomic Commits
*   Branching & Merging
*   History Traversal
*   Blob Deduplication

### › Multi-Tenant Isolation
Strict logical separation of workspaces. Data leakage is prevented at the repository query level, ensuring enterprise-grade security for concurrent tenants.

### › Reality Mandate
*   **No Mocks**: All tests run against real infrastructure.
*   **No Simulation**: Deployments write to real disk paths.
*   **Linux First**: Optimized for systemd, ext4, and standard Unix permissions.

---

## ■ GETTING STARTED

### › Prerequisites

*   Go 1.21+
*   PostgreSQL 15 running on port `5432`
*   Redis 7 running on port `6379`

### › Configuration

Set the following environment variables:

```bash
export DB_DSN="postgres://user:pass@localhost:5432/openbook?sslmode=disable"
export REDIS_ADDR="localhost:6379"
export APP_PORT="8080"
export STORAGE_PATH="/var/lib/openbook/sites"
export JWT_SECRET="your-256-bit-secret"
```

### › Local Development

1.  **Clone Repository**
    ```bash
    git clone https://github.com/rsdenck/openbook.git
    cd openbook
    ```

2.  **Run Migrations**
    ```bash
    migrate -path ./migrations -database "$DB_DSN" up
    ```

3.  **Start Services**
    ```bash
    # Terminal 1: API
    go run ./cmd/api

    # Terminal 2: Worker
    go run ./cmd/worker
    ```

---

## ■ API REFERENCE

### › Deployments

`POST /api/v1/deployments`

Triggers a new deployment for a specific site and commit.

**Payload:**
```json
{
  "site_id": "uuid",
  "commit_hash": "sha256",
  "environment": "production"
}
```

### › Git Operations

`POST /api/v1/branches`
Creates a new branch pointer.

`POST /api/v1/merge`
Performs a fast-forward or merge commit between branches.

---

## ■ CI/CD PIPELINE

The project utilizes a comprehensive **GitHub Actions** workflow enforcing quality gates:

1.  **Validation**: `go vet`, `staticcheck`
2.  **Integration**: Full lifecycle tests against service containers.
3.  **Build**: Compilation of static binaries for Linux `amd64`.
4.  **Release**: Semantic version tagging.
5.  **Deploy**: SCP and SSH execution to production servers.

---

**© 2026 OPENBOOK PROJECT** | EXECUTE. DEPLOY. SURVIVE.
