-- ============================================================================
-- Migration: Create project_members table
-- Tracks which users are assigned to which projects (many-to-many).
-- ============================================================================

CREATE TABLE IF NOT EXISTS project_members (
    -- Composite primary key: one membership per user per project
    project_id  UUID NOT NULL,
    user_id     UUID NOT NULL,

    -- The role this user plays in this specific project (e.g., "developer", "qa")
    role        VARCHAR(50) NOT NULL DEFAULT 'member',

    -- When the user was assigned
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Constraints
    PRIMARY KEY (project_id, user_id),
    CONSTRAINT fk_pm_project FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    CONSTRAINT fk_pm_user    FOREIGN KEY (user_id)    REFERENCES registrations(id) ON DELETE CASCADE
);

-- Index for the reverse lookup: "which projects is this user in?"
CREATE INDEX IF NOT EXISTS idx_project_members_user_id ON project_members(user_id);
