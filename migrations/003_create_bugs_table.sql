-- ============================================================================
-- Migration: Create bugs table
-- Tracks bug reports within projects.
-- Run this SQL in your Supabase Dashboard > SQL Editor
-- ============================================================================

CREATE TABLE IF NOT EXISTS bugs (
    -- Primary key: auto-generated UUID
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Project this bug belongs to
    project_id   UUID NOT NULL,

    -- Human-readable bug identifier (e.g., "BUG-1", "BUG-42")
    bug_number   VARCHAR(20) NOT NULL,

    -- Core bug fields (sent by the client)
    title        VARCHAR(255) NOT NULL,
    priority     VARCHAR(10) NOT NULL DEFAULT 'medium',  -- critical, high, medium, low
    description  TEXT,                                    -- Detailed bug description
    steps        TEXT[] DEFAULT '{}',                     -- Reproduction steps
    version      VARCHAR(50),                             -- Build number e.g. "v1.2.4 (203)"
    platform     VARCHAR(100),                            -- e.g. "Mobile Safari", "Chrome"

    -- Server-managed fields
    status       VARCHAR(20) NOT NULL DEFAULT 'open',     -- open, in_progress, resolved, closed
    created_by   UUID NOT NULL,                           -- Reporter (FK to registrations)
    assigned_to  UUID,                                    -- Assignee (FK to registrations, optional)

    -- Timestamps
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT fk_bug_project   FOREIGN KEY (project_id)  REFERENCES projects(id) ON DELETE CASCADE,
    CONSTRAINT fk_bug_creator   FOREIGN KEY (created_by)  REFERENCES registrations(id),
    CONSTRAINT fk_bug_assignee  FOREIGN KEY (assigned_to) REFERENCES registrations(id),
    CONSTRAINT chk_bug_priority CHECK (priority IN ('critical', 'high', 'medium', 'low')),
    CONSTRAINT chk_bug_status   CHECK (status IN ('open', 'in_progress', 'resolved', 'closed')),
    CONSTRAINT uq_bug_number_project UNIQUE (project_id, bug_number)
);

-- Indexes for common query patterns
CREATE INDEX idx_bugs_project_id  ON bugs(project_id);   -- List bugs by project
CREATE INDEX idx_bugs_created_by  ON bugs(created_by);    -- Bugs reported by a user
CREATE INDEX idx_bugs_assigned_to ON bugs(assigned_to);   -- Bugs assigned to a user
CREATE INDEX idx_bugs_status      ON bugs(status);        -- Filter bugs by status
