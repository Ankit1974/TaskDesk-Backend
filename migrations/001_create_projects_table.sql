-- ============================================================================
-- Migration: Create projects table
-- Run this SQL in your Supabase Dashboard > SQL Editor
-- ============================================================================

-- Projects table stores all project records created by Project Managers (PMs).
-- Each project is linked to a registration record via the created_by foreign key.
CREATE TABLE IF NOT EXISTS projects (
    -- Primary key: auto-generated UUID
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Core project fields (sent by the client)
    project_name    VARCHAR(255) NOT NULL,                  -- Name of the project
    description     TEXT NOT NULL,                           -- Detailed project scope and goals
    icon            VARCHAR(50) NOT NULL DEFAULT 'language', -- Material Icon name for the project
    teams           TEXT[] DEFAULT '{}',                     -- Assigned team keys (e.g., {"backend", "frontend"})
    start_date      DATE,                                   -- Optional planned start date

    -- Server-managed fields
    status          VARCHAR(20) NOT NULL DEFAULT 'planning', -- Current status: active, planning, on_hold, completed
    workspace_id    VARCHAR(20) NOT NULL,                    -- Auto-generated short ID (e.g., "E-C-K1R2")
    created_by      UUID NOT NULL,                           -- FK to registrations.id (the PM who created it)
    progress        INTEGER NOT NULL DEFAULT 0,              -- Completion percentage (0-100)
    member_count    INTEGER NOT NULL DEFAULT 0,              -- Number of team members assigned

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),      -- Record creation time
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),      -- Last modification time

    -- Constraints
    CONSTRAINT fk_created_by FOREIGN KEY (created_by) REFERENCES registrations(id),
    CONSTRAINT chk_status    CHECK (status IN ('active', 'planning', 'on_hold', 'completed')),
    CONSTRAINT chk_icon      CHECK (icon IN ('language', 'smartphone', 'cloud', 'storage', 'cloud-upload')),
    CONSTRAINT chk_progress  CHECK (progress >= 0 AND progress <= 100)
);

-- Indexes for common query patterns
CREATE INDEX IF NOT EXISTS idx_projects_created_by ON projects(created_by);  -- Filter projects by creator
CREATE INDEX IF NOT EXISTS idx_projects_status     ON projects(status);      -- Filter projects by status
