-- =================================================================
-- Facet Database Schema — Initial Migration
-- =================================================================
-- This script creates the complete database schema for the Facet
-- feature flag management platform.
-- All tables are created in the "facet" schema.
-- =================================================================

CREATE SCHEMA IF NOT EXISTS facet;

-- =================================================================
-- Custom Enum Types
-- =================================================================
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type t JOIN pg_namespace n ON t.typnamespace = n.oid WHERE t.typname = 'user_role' AND n.nspname = 'facet') THEN
        CREATE TYPE facet.user_role AS ENUM ('system_admin', 'tenant_super_admin', 'tenant_admin', 'tenant_viewer');
    END IF;
END $$;

DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type t JOIN pg_namespace n ON t.typnamespace = n.oid WHERE t.typname = 'tenant_status' AND n.nspname = 'facet') THEN
        CREATE TYPE facet.tenant_status AS ENUM ('active', 'suspended', 'archived');
    END IF;
END $$;

DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type t JOIN pg_namespace n ON t.typnamespace = n.oid WHERE t.typname = 'notification_level' AND n.nspname = 'facet') THEN
        CREATE TYPE facet.notification_level AS ENUM ('normal', 'important', 'urgent');
    END IF;
END $$;

DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type t JOIN pg_namespace n ON t.typnamespace = n.oid WHERE t.typname = 'notification_type' AND n.nspname = 'facet') THEN
        CREATE TYPE facet.notification_type AS ENUM ('announcement', 'product_update', 'security_alert');
    END IF;
END $$;

DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type t JOIN pg_namespace n ON t.typnamespace = n.oid WHERE t.typname = 'segment_match_type' AND n.nspname = 'facet') THEN
        CREATE TYPE facet.segment_match_type AS ENUM ('ALL', 'ANY');
    END IF;
END $$;

-- =================================================================
-- Tenant Levels Table (plans/pricing tiers)
-- =================================================================
CREATE TABLE IF NOT EXISTS facet.tenant_levels (
    id                 VARCHAR(50)   PRIMARY KEY,
    code               VARCHAR(50)   NOT NULL UNIQUE,
    name               VARCHAR(100)  NOT NULL,
    max_environments   INT           NOT NULL,
    max_feature_flags  INT           NOT NULL,
    max_tenant_admins  INT           NOT NULL,
    max_tenant_viewers INT           NOT NULL,
    allow_sso          BOOLEAN       NOT NULL DEFAULT FALSE,
    allow_api_access   BOOLEAN       NOT NULL DEFAULT FALSE,
    has_audit_log      BOOLEAN       NOT NULL DEFAULT FALSE,
    priority_support   BOOLEAN       NOT NULL DEFAULT FALSE,
    price_monthly      DECIMAL(10,2) NOT NULL DEFAULT 0,
    price_yearly       DECIMAL(10,2) NOT NULL DEFAULT 0,
    is_public          BOOLEAN       NOT NULL DEFAULT FALSE,
    is_active          BOOLEAN       NOT NULL DEFAULT TRUE,
    created_at         TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

-- =================================================================
-- Tenants Table
-- =================================================================
CREATE TABLE IF NOT EXISTS facet.tenants (
    id          VARCHAR(50)        PRIMARY KEY,
    level_id    VARCHAR(50)        NOT NULL DEFAULT 'free',
    name        VARCHAR(255)       NOT NULL,
    description TEXT,
    website     VARCHAR(255),
    status      facet.tenant_status NOT NULL DEFAULT 'active',
    created_by  VARCHAR(50),
    updated_by  VARCHAR(50),
    created_at  TIMESTAMPTZ        NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ        NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

-- =================================================================
-- Users Table
-- =================================================================
CREATE TABLE IF NOT EXISTS facet.users (
    id            VARCHAR(50)   PRIMARY KEY,
    tenant_id     VARCHAR(50),
    role          facet.user_role NOT NULL,
    email         VARCHAR(255)  NOT NULL,
    password_hash VARCHAR(255)  NOT NULL,
    display_name  VARCHAR(100),
    last_login_at TIMESTAMPTZ,
    created_by    VARCHAR(50),
    updated_by    VARCHAR(50),
    created_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_users_tenant_id ON facet.users (tenant_id);
CREATE INDEX IF NOT EXISTS idx_users_email ON facet.users (email);

-- =================================================================
-- Environments Table
-- =================================================================
CREATE TABLE IF NOT EXISTS facet.environments (
    id          VARCHAR(50)  PRIMARY KEY,
    tenant_id   VARCHAR(50)  NOT NULL,
    name        VARCHAR(100) NOT NULL,
    description TEXT,
    sdk_key     VARCHAR(255) NOT NULL,
    is_active   BOOLEAN      NOT NULL DEFAULT TRUE,
    created_by  VARCHAR(50),
    updated_by  VARCHAR(50),
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_environments_tenant_id ON facet.environments (tenant_id);
CREATE INDEX IF NOT EXISTS idx_environments_sdk_key ON facet.environments (sdk_key);

-- =================================================================
-- Feature Flags Table
-- =================================================================
CREATE TABLE IF NOT EXISTS facet.feature_flags (
    id                 VARCHAR(50)  PRIMARY KEY,
    tenant_id          VARCHAR(50)  NOT NULL,
    environment_id     VARCHAR(50)  NOT NULL,
    name               VARCHAR(255) NOT NULL,
    description        TEXT,
    enabled            BOOLEAN      NOT NULL DEFAULT FALSE,
    flag_type          VARCHAR(50)  NOT NULL DEFAULT 'RELEASE',
    rollout_percentage INT          NOT NULL DEFAULT 0,
    expires_at         VARCHAR(100),
    archived_at        TIMESTAMPTZ,
    created_by         VARCHAR(50),
    updated_by         VARCHAR(50),
    created_at         TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at         TIMESTAMPTZ,
    UNIQUE (environment_id, name)
);

CREATE INDEX IF NOT EXISTS idx_feature_flags_tenant_id ON facet.feature_flags (tenant_id);
CREATE INDEX IF NOT EXISTS idx_feature_flags_env_id_name ON facet.feature_flags (environment_id, name);
CREATE INDEX IF NOT EXISTS idx_feature_flags_rollout ON facet.feature_flags (rollout_percentage) WHERE rollout_percentage > 0;
CREATE INDEX IF NOT EXISTS idx_feature_flags_archived ON facet.feature_flags (archived_at) WHERE archived_at IS NOT NULL;

-- =================================================================
-- Feature Flag Whitelist Rules Table
-- =================================================================
CREATE TABLE IF NOT EXISTS facet.feature_flags_rules (
    id              VARCHAR(50) PRIMARY KEY,
    feature_flag_id VARCHAR(50) NOT NULL,
    account_id      VARCHAR(255),
    user_ids        TEXT[]      NOT NULL DEFAULT '{}',
    created_by      VARCHAR(50),
    updated_by      VARCHAR(50),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ff_rules_feature_flag_id ON facet.feature_flags_rules (feature_flag_id);

-- =================================================================
-- Feature Flag History Table
-- =================================================================
CREATE TABLE IF NOT EXISTS facet.feature_flag_history (
    history_id        VARCHAR(50)  PRIMARY KEY,
    feature_flag_id   VARCHAR(50)  NOT NULL,
    change_time       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    change_type       VARCHAR(50)  NOT NULL,
    changed_by        VARCHAR(50)  NOT NULL,
    name              VARCHAR(255) NOT NULL,
    description       TEXT         NOT NULL DEFAULT '',
    enabled           BOOLEAN      NOT NULL DEFAULT FALSE,
    flag_type         VARCHAR(50)  NOT NULL DEFAULT '',
    rollout_percentage INT         NOT NULL DEFAULT 0,
    expires_at        VARCHAR(100),
    rules_snapshot    JSONB        NOT NULL DEFAULT '[]'
);

CREATE INDEX IF NOT EXISTS idx_ff_history_feature_flag_id ON facet.feature_flag_history (feature_flag_id);

-- =================================================================
-- Segments Table (user segments for attribute-based targeting)
-- =================================================================
CREATE TABLE IF NOT EXISTS facet.segments (
    id          VARCHAR(50) PRIMARY KEY,
    tenant_id   VARCHAR(50) NOT NULL,
    name        VARCHAR(255) NOT NULL,
    description TEXT DEFAULT '',
    match_type  VARCHAR(20) NOT NULL DEFAULT 'ANY',
    created_by  VARCHAR(50),
    updated_by  VARCHAR(50),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_segments_tenant_id ON facet.segments (tenant_id) WHERE deleted_at IS NULL;

-- =================================================================
-- Segment Conditions Table
-- =================================================================
CREATE TABLE IF NOT EXISTS facet.segment_conditions (
    id            VARCHAR(50) PRIMARY KEY,
    segment_id    VARCHAR(50) NOT NULL REFERENCES facet.segments(id) ON DELETE CASCADE,
    attribute_key VARCHAR(255) NOT NULL,
    operator      VARCHAR(50) NOT NULL DEFAULT 'equals',
    value         VARCHAR(500) NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_segment_conditions_segment_id ON facet.segment_conditions (segment_id);

-- =================================================================
-- Feature Flag ↔ Segment Association Table
-- =================================================================
CREATE TABLE IF NOT EXISTS facet.feature_flag_segments (
    feature_flag_id VARCHAR(50) NOT NULL REFERENCES facet.feature_flags(id) ON DELETE CASCADE,
    segment_id      VARCHAR(50) NOT NULL REFERENCES facet.segments(id) ON DELETE CASCADE,
    PRIMARY KEY (feature_flag_id, segment_id)
);

CREATE INDEX IF NOT EXISTS idx_feature_flag_segments_flag_id ON facet.feature_flag_segments (feature_flag_id);

-- =================================================================
-- System Notifications Table
-- =================================================================
CREATE TABLE IF NOT EXISTS facet.system_notifications (
    id         VARCHAR(50)           PRIMARY KEY,
    level      facet.notification_level NOT NULL DEFAULT 'normal',
    type       facet.notification_type  NOT NULL DEFAULT 'announcement',
    title      VARCHAR(255)          NOT NULL,
    content    TEXT                  NOT NULL,
    deleted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ           NOT NULL DEFAULT NOW()
);

-- =================================================================
-- User Notifications Table
-- =================================================================
CREATE TABLE IF NOT EXISTS facet.user_notifications (
    id              VARCHAR(50) PRIMARY KEY,
    tenant_id       VARCHAR(50) NOT NULL,
    user_id         VARCHAR(50) NOT NULL,
    notification_id VARCHAR(50) NOT NULL,
    is_read         BOOLEAN     NOT NULL DEFAULT FALSE,
    read_at         TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, notification_id)
);

CREATE INDEX IF NOT EXISTS idx_user_notifications_user_id ON facet.user_notifications (user_id);

-- =================================================================
-- Evaluation Logs Table (SDK evaluation event history)
-- =================================================================
CREATE TABLE IF NOT EXISTS facet.evaluation_logs (
    event_id               VARCHAR(50)  PRIMARY KEY,
    event_time             TIMESTAMPTZ  NOT NULL,
    sdk_key                VARCHAR(255) NOT NULL,
    feature_name           VARCHAR(255) NOT NULL,
    account_id             VARCHAR(255),
    user_id                VARCHAR(255),
    app_name               VARCHAR(255),
    matched_result         BOOLEAN      NOT NULL,
    evaluation_duration_ms BIGINT       NOT NULL DEFAULT 0,
    error_message          TEXT,
    client_ip              VARCHAR(50),
    user_agent             TEXT,
    host                   VARCHAR(255),
    proto                  VARCHAR(20),
    referer                TEXT
);

CREATE INDEX IF NOT EXISTS idx_evaluation_logs_event_time ON facet.evaluation_logs (event_time);
CREATE INDEX IF NOT EXISTS idx_evaluation_logs_sdk_key ON facet.evaluation_logs (sdk_key);
CREATE INDEX IF NOT EXISTS idx_evaluation_logs_feature_name ON facet.evaluation_logs (feature_name);

-- =================================================================
-- Seed Data
-- =================================================================

-- Tenant levels (plans)
INSERT INTO facet.tenant_levels (id, code, name, max_environments, max_feature_flags, max_tenant_admins, max_tenant_viewers, allow_sso, allow_api_access, has_audit_log, priority_support, price_monthly, price_yearly, is_public, is_active)
VALUES
    ('free',       'free',       'Free',       2,  10,  1,  3,  FALSE, FALSE, FALSE, FALSE, 0,   0,   TRUE,  TRUE),
    ('pro',        'pro',        'Pro',        10, 100, 10, 20, TRUE,  TRUE,  TRUE,  FALSE, 99,  999,  TRUE,  TRUE),
    ('enterprise', 'enterprise', 'Enterprise', -1, -1,  -1, -1,  TRUE,  TRUE,  TRUE,  TRUE,  499, 4999, TRUE,  TRUE)
ON CONFLICT (id) DO NOTHING;