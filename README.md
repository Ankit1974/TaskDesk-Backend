# TaskDesk Backend — API Documentation

> **Version:** 1.0.0 &nbsp;|&nbsp; **Base URL:** `http://localhost:8080/api/v1` &nbsp;|&nbsp; **Last Updated:** 2026-02-14

---

## Table of Contents

1. [Architecture Overview](#1-architecture-overview)
2. [System Architecture Diagram](#2-system-architecture-diagram)
3. [Project Structure](#3-project-structure)
4. [Tech Stack](#4-tech-stack)
5. [Environment Setup](#5-environment-setup)
6. [Database Schema](#6-database-schema)
7. [Authentication & Authorization](#7-authentication--authorization)
8. [API Reference](#8-api-reference)
   - [Health Check](#81-health-check)
   - [User Registration](#82-user-registration)
   - [Create Project](#83-create-project)
9. [Error Handling](#9-error-handling)
10. [Request Lifecycle](#10-request-lifecycle)

---

## 1. Architecture Overview

TaskDesk Backend follows a **layered architecture** built with Go and the Gin framework. Every incoming HTTP request flows through a well-defined pipeline of middleware, validation, business logic, and persistence layers.

```
┌──────────────────────────────────────────────────────────────────┐
│                        CLIENT (Browser / Mobile)                 │
└────────────────────────────────┬─────────────────────────────────┘
                                 │  HTTPS
                                 ▼
┌──────────────────────────────────────────────────────────────────┐
│                     GIN HTTP SERVER (:8080)                       │
│  ┌────────────┐  ┌────────────────┐  ┌────────────────────────┐  │
│  │  Recovery   │→│  Logger        │→│  Route Matching         │  │
│  │  Middleware  │  │  Middleware    │  │  /api/v1/*             │  │
│  └────────────┘  └────────────────┘  └──────────┬─────────────┘  │
│                                                  │                │
│          ┌───────────────────────────────────────┤                │
│          │ Public Routes                         │ Protected      │
│          │                                       │ Routes         │
│          ▼                                       ▼                │
│  ┌──────────────┐                   ┌────────────────────────┐   │
│  │  Handler     │                   │  Auth Middleware        │   │
│  │  (no auth)   │                   │  (JWT Validation)       │   │
│  └──────┬───────┘                   └──────────┬─────────────┘   │
│         │                                      │                  │
│         │                                      ▼                  │
│         │                           ┌────────────────────────┐   │
│         │                           │  Role Middleware        │   │
│         │                           │  (RBAC Check)           │   │
│         │                           └──────────┬─────────────┘   │
│         │                                      │                  │
│         ▼                                      ▼                  │
│  ┌─────────────────────────────────────────────────────────────┐  │
│  │                    REQUEST HANDLER                           │  │
│  │  1. Bind & Validate JSON body                               │  │
│  │  2. Execute Business Logic                                  │  │
│  │  3. Database Query (5s timeout)                             │  │
│  │  4. Return JSON Response                                    │  │
│  └──────────────────────────┬──────────────────────────────────┘  │
└─────────────────────────────┼────────────────────────────────────┘
                              │
                              ▼
┌──────────────────────────────────────────────────────────────────┐
│                   POSTGRESQL (Supabase)                           │
│  ┌─────────────────┐       ┌──────────────────────┐              │
│  │  registrations   │◄─────│  projects             │              │
│  │  table           │  FK  │  table                │              │
│  └─────────────────┘       └──────────────────────┘              │
└──────────────────────────────────────────────────────────────────┘
```

---

## 2. System Architecture Diagram

```mermaid
graph TB
    subgraph Client
        A[Browser / Mobile App]
    end

    subgraph "Gin HTTP Server :8080"
        B[Recovery + Logger Middleware]
        C{Route Matcher}
        D[Auth Middleware<br/>JWT Validation]
        E[Role Middleware<br/>RBAC]
        F[HealthCheck Handler]
        G[Register Handler]
        H[CreateProject Handler]
    end

    subgraph "Data Layer"
        I[(PostgreSQL<br/>Supabase)]
    end

    subgraph "External Services"
        J[Supabase Auth<br/>JWT Issuer]
    end

    A -->|HTTP Request| B
    B --> C
    C -->|GET /health| F
    C -->|POST /register| G
    C -->|POST /projects| D
    D -->|Valid Token| E
    D -.->|Invalid Token| A
    E -->|Role = PM| H
    E -.->|Role ≠ PM| A
    F -->|Ping| I
    G -->|INSERT| I
    H -->|INSERT| I
    J -.->|Issues JWT| A
    D -.->|Verifies JWT using<br/>SUPABASE_JWT_SECRET| J

    style D fill:#f9d71c,stroke:#333
    style E fill:#f9d71c,stroke:#333
    style I fill:#336791,stroke:#333,color:#fff
```

---

## 3. Project Structure

```
TaskDesk-Backend/
│
├── cmd/
│   └── api/
│       └── main.go                  # Application entry point
│
├── internal/
│   ├── api/
│   │   ├── handlers/
│   │   │   ├── health.go            # GET  /health
│   │   │   ├── registration.go      # POST /register
│   │   │   └── project.go           # POST /projects
│   │   ├── middleware/
│   │   │   └── auth.go              # JWT auth + role-based access control
│   │   └── router/
│   │       └── router.go            # Route definitions & grouping
│   │
│   ├── config/
│   │   └── config.go                # Environment variable loading (Viper)
│   │
│   ├── db/
│   │   └── db.go                    # PostgreSQL connection pool (pgx)
│   │
│   ├── logger/
│   │   └── logger.go                # Structured logging (Zap)
│   │
│   └── model/
│       ├── registration.go          # Registration struct + validation tags
│       └── project.go               # Project & CreateProjectRequest structs
│
├── migrations/
│   └── 001_create_projects_table.sql
│
├── .env.example                     # Environment variable template
├── go.mod
└── go.sum
```

---

## 4. Tech Stack

| Layer            | Technology                          |
|------------------|-------------------------------------|
| Language         | Go 1.24+                            |
| HTTP Framework   | Gin v1.10                           |
| Database         | PostgreSQL (hosted on Supabase)     |
| DB Driver        | pgx/v5 (connection pooling)         |
| Authentication   | Supabase Auth (JWT / HMAC-SHA256)   |
| Configuration    | Viper                               |
| Logging          | Uber Zap (structured logging)       |
| Validation       | Gin Binding Tags (struct-level)     |

---

## 5. Environment Setup

Copy `.env.example` to `.env` and configure:

```env
# ─── Server ────────────────────────────────────
APP_PORT=8080                    # Port the server listens on
ENV=development                  # "development" | "production"

# ─── Supabase ──────────────────────────────────
SUPABASE_URL=https://xxx.supabase.co
SUPABASE_ANON_KEY=eyJhbGci...    # Supabase anon/public key
SUPABASE_JWT_SECRET=your-jwt-... # Used to verify JWT signatures

# ─── Database ──────────────────────────────────
DATABASE_URL=postgresql://user:pass@host:6543/postgres
```

**Start the server:**

```bash
go run cmd/api/main.go
```

**Startup Sequence:**

```
LoadConfig() → InitLogger() → InitDB() → SetupRouter() → Run(:8080)
      │              │              │             │
      ▼              ▼              ▼             ▼
  Read .env     Zap logger     pgx pool     Gin engine
  via Viper     (dev/prod)     + ping DB     + routes
```

---

## 6. Database Schema

### Entity-Relationship Diagram

```mermaid
erDiagram
    REGISTRATIONS {
        UUID id PK "gen_random_uuid()"
        VARCHAR full_name "NOT NULL"
        VARCHAR email "UNIQUE, NOT NULL"
        TEXT organisation_name "NOT NULL"
        VARCHAR role "NOT NULL (e.g. PM, Developer)"
        TIMESTAMPTZ created_at "DEFAULT NOW()"
    }

    PROJECTS {
        UUID id PK "gen_random_uuid()"
        VARCHAR project_name "NOT NULL"
        TEXT description "NOT NULL"
        VARCHAR icon "DEFAULT 'language'"
        TEXT_ARRAY teams "DEFAULT '{}'"
        DATE start_date "NULLABLE"
        VARCHAR status "DEFAULT 'planning'"
        VARCHAR workspace_id "UNIQUE, NOT NULL"
        UUID created_by FK "NOT NULL"
        INTEGER progress "DEFAULT 0, 0-100"
        INTEGER member_count "DEFAULT 0"
        TIMESTAMPTZ created_at "DEFAULT NOW()"
        TIMESTAMPTZ updated_at "DEFAULT NOW()"
    }

    REGISTRATIONS ||--o{ PROJECTS : "created_by"
```

### Constraints

| Table    | Constraint       | Rule                                                        |
|----------|------------------|-------------------------------------------------------------|
| projects | `chk_status`     | `status IN ('active', 'planning', 'on_hold', 'completed')` |
| projects | `chk_icon`       | `icon IN ('language', 'smartphone', 'cloud', 'storage', 'cloud-upload')` |
| projects | `chk_progress`   | `progress >= 0 AND progress <= 100`                         |
| projects | `fk_created_by`  | `FOREIGN KEY (created_by) REFERENCES registrations(id)`     |

### Indexes

| Index                       | Column       | Purpose                     |
|-----------------------------|-------------|-----------------------------|
| `idx_projects_created_by`   | `created_by` | Filter projects by creator  |
| `idx_projects_status`       | `status`     | Filter projects by status   |

---

## 7. Authentication & Authorization

TaskDesk uses **Supabase Auth** for issuing JWTs and a custom middleware stack for validation and role-based access control.

### Authentication Flow

```mermaid
sequenceDiagram
    participant C as Client
    participant S as Supabase Auth
    participant M as Auth Middleware
    participant R as RequireRole Middleware
    participant H as Handler
    participant DB as PostgreSQL

    C->>S: Login (email/password)
    S-->>C: JWT (access_token)

    C->>M: POST /api/v1/projects<br/>Authorization: Bearer <token>

    M->>M: Extract Bearer token
    M->>M: Parse JWT with SUPABASE_JWT_SECRET
    M->>M: Validate signature (HMAC-SHA256)
    M->>M: Check token expiration

    alt Token Invalid
        M-->>C: 401 Unauthorized
    end

    M->>M: Extract claims (sub, email)
    M->>DB: SELECT id, role FROM registrations WHERE email = ?

    alt User Not Found
        M-->>C: 401 "User not registered"
    end

    M->>M: Build UserContext{SupabaseUserID, Email, RegistrationID, Role}
    M->>R: Pass request with UserContext

    R->>R: Check user.Role against allowed roles (case-insensitive)

    alt Role Mismatch
        R-->>C: 403 "Insufficient permissions"
    end

    R->>H: Authorized — proceed to handler
    H->>DB: Execute business logic
    H-->>C: 201 Created (JSON response)
```

### UserContext Object

After successful authentication, the following user context is injected into the request:

```go
type UserContext struct {
    SupabaseUserID string   // From JWT "sub" claim
    Email          string   // From JWT "email" claim
    RegistrationID string   // From registrations table (PK)
    Role           string   // From registrations table (e.g. "PM")
}
```

### Route Protection Matrix

| Endpoint               | Auth Required | Role Required | Middleware Chain                          |
|------------------------|:------------:|:------------:|-------------------------------------------|
| `GET  /api/v1/health`  | No           | —            | `Router → Handler`                        |
| `POST /api/v1/register`| No           | —            | `Router → Handler`                        |
| `POST /api/v1/projects`| Yes          | PM           | `Router → AuthMiddleware → RequireRole("PM") → Handler` |

---

## 8. API Reference

---

### 8.1 Health Check

Check server and database connectivity.

| Property       | Value                   |
|----------------|-------------------------|
| **Endpoint**   | `GET /api/v1/health`    |
| **Auth**       | None                    |
| **Rate Limit** | —                       |

#### Request

```
GET /api/v1/health
```

No headers, body, or parameters required.

#### Flow Diagram

```mermaid
flowchart LR
    A[GET /health] --> B[Handler]
    B --> C{Ping DB}
    C -->|Success| D["db_status: up"]
    C -->|Failure| E["db_status: down"]
    D --> F["200 OK"]
    E --> F
```

#### Response

**`200 OK`**

```json
{
  "status": "up",
  "db_status": "up"
}
```

> `db_status` will be `"down"` if the database connection pool fails to respond.

---

### 8.2 User Registration

Register a new user with their organisation and role.

| Property       | Value                      |
|----------------|----------------------------|
| **Endpoint**   | `POST /api/v1/register`    |
| **Auth**       | None                       |
| **Content-Type** | `application/json`       |

#### Request Body

| Field               | Type   | Required | Validation          | Description                  |
|---------------------|--------|:--------:|---------------------|------------------------------|
| `full_name`         | string | Yes      | non-empty           | User's full name             |
| `email`             | string | Yes      | valid email format  | User's email address         |
| `organisation_name` | string | Yes      | non-empty           | Organisation the user belongs to |
| `role`              | string | Yes      | non-empty           | User role (e.g. `PM`, `Developer`) |

#### Example Request

```bash
curl -X POST http://localhost:8080/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{
    "full_name": "Ankit Kumar",
    "email": "ankit@taskdesk.io",
    "organisation_name": "TaskDesk Inc.",
    "role": "PM"
  }'
```

#### Flow Diagram

```mermaid
flowchart TD
    A["POST /register"] --> B["Bind JSON to Registration struct"]
    B --> C{Validation<br/>passed?}
    C -->|No| D["400 Bad Request<br/>(validation details)"]
    C -->|Yes| E["Create 5s timeout context"]
    E --> F["INSERT INTO registrations<br/>(full_name, email, org, role)<br/>RETURNING id, created_at"]
    F --> G{DB query<br/>succeeded?}
    G -->|No| H["500 Internal Server Error<br/>'Failed to save registration'"]
    G -->|Yes| I["201 Created<br/>(full registration object)"]

    style D fill:#e74c3c,color:#fff
    style H fill:#e74c3c,color:#fff
    style I fill:#27ae60,color:#fff
```

#### Responses

**`201 Created`** — Registration successful

```json
{
  "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "full_name": "Ankit Kumar",
  "email": "ankit@taskdesk.io",
  "organisation_name": "TaskDesk Inc.",
  "role": "PM",
  "created_at": "2026-02-14T10:30:00Z"
}
```

**`400 Bad Request`** — Validation failed

```json
{
  "error": "Key: 'Registration.Email' Error:Field validation for 'Email' failed on the 'email' tag"
}
```

---

### 8.3 Create Project

Create a new project. Restricted to authenticated users with the **PM** role.

| Property         | Value                      |
|------------------|----------------------------|
| **Endpoint**     | `POST /api/v1/projects`    |
| **Auth**         | Bearer Token (Supabase JWT)|
| **Required Role**| `PM`                       |
| **Content-Type** | `application/json`         |

#### Request Headers

```
Authorization: Bearer <supabase-jwt-token>
Content-Type: application/json
```

#### Request Body

| Field          | Type     | Required | Validation                                           | Description                    |
|----------------|----------|:--------:|------------------------------------------------------|--------------------------------|
| `project_name` | string   | Yes      | non-empty                                            | Name of the project            |
| `description`  | string   | Yes      | non-empty                                            | Project scope and goals        |
| `teams`        | string[] | Yes      | each must be: `backend` `frontend` `mobile` `qa` `uiux` | Assigned team categories       |
| `start_date`   | string   | No       | format: `YYYY-MM-DD`                                 | Planned start date             |

#### Example Request

```bash
curl -X POST http://localhost:8080/api/v1/projects \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..." \
  -H "Content-Type: application/json" \
  -d '{
    "project_name": "Mobile App Redesign",
    "description": "Complete redesign of the mobile application UI/UX",
    "teams": ["frontend", "mobile", "uiux"],
    "start_date": "2026-03-01"
  }'
```

#### Flow Diagram

```mermaid
flowchart TD
    A["POST /projects"] --> B["Auth Middleware"]

    B --> B1["Extract Bearer token"]
    B1 --> B2["Parse & validate JWT<br/>(HMAC-SHA256)"]
    B2 --> B3{Token valid?}
    B3 -->|No| B4["401 Unauthorized"]
    B3 -->|Yes| B5["Extract claims: sub, email"]
    B5 --> B6["SELECT id, role FROM registrations<br/>WHERE email = ?"]
    B6 --> B7{User found?}
    B7 -->|No| B8["401 'User not registered'"]
    B7 -->|Yes| B9["Build UserContext"]

    B9 --> C["Role Middleware"]
    C --> C1{Role = PM?}
    C1 -->|No| C2["403 'Insufficient permissions'"]
    C1 -->|Yes| D["CreateProject Handler"]

    D --> D1["Bind & validate JSON body"]
    D1 --> D2{Validation<br/>passed?}
    D2 -->|No| D3["400 Bad Request"]
    D2 -->|Yes| D4{start_date<br/>provided?}

    D4 -->|Yes| D5["Parse YYYY-MM-DD"]
    D5 --> D6{Format valid?}
    D6 -->|No| D7["400 'Invalid start_date format'"]
    D6 -->|Yes| D8["Generate workspace_id"]
    D4 -->|No| D8

    D8 --> D9["INSERT INTO projects<br/>(name, desc, teams, date,<br/>status='planning',<br/>workspace_id, created_by)<br/>RETURNING id, status,<br/>progress, member_count,<br/>created_at, updated_at"]

    D9 --> D10{DB query<br/>succeeded?}
    D10 -->|No| D11["500 'Failed to create project'"]
    D10 -->|Yes| D12["201 Created<br/>(full project object)"]

    style B4 fill:#e74c3c,color:#fff
    style B8 fill:#e74c3c,color:#fff
    style C2 fill:#e74c3c,color:#fff
    style D3 fill:#e74c3c,color:#fff
    style D7 fill:#e74c3c,color:#fff
    style D11 fill:#e74c3c,color:#fff
    style D12 fill:#27ae60,color:#fff
```

#### Workspace ID Generation

The system auto-generates a short, human-readable workspace identifier:

```
Format:  [PREFIX]-[SUFFIX]

PREFIX = first 3 chars of project_name (UPPERCASED)
SUFFIX = last 4 chars of Unix timestamp in base36

Example: "Mobile App Redesign" → "MOB-K1R2"
```

#### Server-Managed Fields

These fields are set automatically and cannot be overridden by the client:

| Field          | Default Value | Description                       |
|----------------|---------------|-----------------------------------|
| `id`           | UUID (auto)   | Primary key                       |
| `status`       | `"planning"`  | Initial project status            |
| `workspace_id` | Generated     | Short unique project identifier   |
| `created_by`   | Auth user ID  | Registration ID of the PM         |
| `progress`     | `0`           | Completion percentage (0–100)     |
| `member_count` | `0`           | Number of assigned members        |
| `created_at`   | `NOW()`       | Record creation timestamp         |
| `updated_at`   | `NOW()`       | Last modification timestamp       |

#### Responses

**`201 Created`** — Project created successfully

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "project_name": "Mobile App Redesign",
  "description": "Complete redesign of the mobile application UI/UX",
  "icon": "",
  "teams": ["frontend", "mobile", "uiux"],
  "start_date": "2026-03-01",
  "status": "planning",
  "workspace_id": "MOB-K1R2",
  "created_by": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "progress": 0,
  "member_count": 0,
  "created_at": "2026-02-14T10:30:00Z",
  "updated_at": "2026-02-14T10:30:00Z"
}
```

**`400 Bad Request`** — Validation failed

```json
{ "error": "Key: 'CreateProjectRequest.Teams' Error:Field validation for 'Teams' failed on the 'oneof' tag" }
```

**`401 Unauthorized`** — Authentication failed

```json
{ "error": "Missing or invalid authorization header" }
```

```json
{ "error": "Invalid or expired token" }
```

```json
{ "error": "User not registered" }
```

**`403 Forbidden`** — Insufficient role

```json
{ "error": "Insufficient permissions. Required role: PM" }
```

**`500 Internal Server Error`** — Database failure

```json
{ "error": "Failed to create project" }
```

---

## 9. Error Handling

All errors follow a consistent JSON envelope:

```json
{ "error": "<descriptive message>" }
```

### Error Code Reference

| HTTP Code | Meaning               | When It Occurs                                                     |
|-----------|-----------------------|--------------------------------------------------------------------|
| `400`     | Bad Request           | Invalid JSON body, missing required fields, invalid enum values, malformed date |
| `401`     | Unauthorized          | Missing `Authorization` header, invalid/expired JWT, unregistered user |
| `403`     | Forbidden             | Authenticated user lacks the required role for the endpoint        |
| `500`     | Internal Server Error | Database query failure, unexpected server error                    |

### Error Handling Pattern in Code

```
Request arrives
    │
    ├─ Validation Error?     → 400 + detailed validation message
    ├─ Auth Error?           → 401 + auth-specific message
    ├─ Role Error?           → 403 + required role message
    ├─ Business Logic Error? → 400 + human-readable message
    └─ Database Error?       → 500 + generic message (details logged server-side)
```

> **Security Note:** Database error details are logged via Zap but never exposed to the client. Only generic messages like `"Failed to create project"` are returned in the response body.

---

## 10. Request Lifecycle

Every request to a protected endpoint goes through the following complete lifecycle:

```mermaid
sequenceDiagram
    participant C as Client
    participant GIN as Gin Engine
    participant AUTH as AuthMiddleware
    participant ROLE as RequireRole
    participant H as Handler
    participant DB as PostgreSQL

    Note over C,DB: ── Request Phase ──
    C->>GIN: HTTP Request
    GIN->>GIN: Recovery Middleware (panic guard)
    GIN->>GIN: Logger Middleware (request log)
    GIN->>GIN: Route Matching

    alt Public Route (/health, /register)
        GIN->>H: Direct to Handler
    else Protected Route (/projects)
        GIN->>AUTH: Pass to AuthMiddleware
        AUTH->>AUTH: Parse Authorization header
        AUTH->>AUTH: Validate JWT (signature + expiry)
        AUTH->>DB: Lookup user by email
        DB-->>AUTH: {id, role}
        AUTH->>AUTH: Build UserContext
        AUTH->>ROLE: Pass to RequireRole
        ROLE->>ROLE: Check role match
        ROLE->>H: Authorized → Handler
    end

    Note over C,DB: ── Processing Phase ──
    H->>H: Bind & validate request body
    H->>H: Execute business logic
    H->>DB: SQL query (5s timeout)
    DB-->>H: Query result

    Note over C,DB: ── Response Phase ──
    H-->>C: JSON Response (2xx / 4xx / 5xx)
```

---

## Appendix

### A. Status Values for Projects

| Status      | Description                              |
|-------------|------------------------------------------|
| `planning`  | Initial state — project is being planned |
| `active`    | Project is currently in progress         |
| `on_hold`   | Project is temporarily paused            |
| `completed` | Project has been finished                |

### B. Valid Team Categories

| Value      | Description           |
|------------|-----------------------|
| `backend`  | Backend development   |
| `frontend` | Frontend development  |
| `mobile`   | Mobile development    |
| `qa`       | Quality assurance     |
| `uiux`     | UI/UX design          |

### C. Valid Icon Values

| Value          | Description                |
|----------------|----------------------------|
| `language`     | Code / programming (default)|
| `smartphone`   | Mobile device              |
| `cloud`        | Cloud service              |
| `storage`      | Storage / database         |
| `cloud-upload` | Cloud upload / deployment  |
