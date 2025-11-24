-- Migration: 005_add_rbac_tables
-- Description: Add tables for RBAC (roles, user_roles)
-- Created: 2024-01-24

-- Create roles table
CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    permissions TEXT[] NOT NULL DEFAULT '{}',
    is_system BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

-- Create index on role name
CREATE INDEX idx_roles_name ON roles(name) WHERE deleted_at IS NULL;

-- Create user_roles junction table
CREATE TABLE IF NOT EXISTS user_roles (
    user_id UUID NOT NULL,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, role_id)
);

-- Create indexes for user_roles
CREATE INDEX idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX idx_user_roles_role_id ON user_roles(role_id);

-- Insert default system roles
INSERT INTO roles (id, name, description, permissions, is_system) VALUES
    (gen_random_uuid(), 'admin', 'Administrator with full access', ARRAY['admin:*'], TRUE),
    (gen_random_uuid(), 'editor', 'Can create and edit workflows', ARRAY[
        'workflow:create', 'workflow:read', 'workflow:update', 'workflow:execute',
        'node:read', 'execution:read'
    ], TRUE),
    (gen_random_uuid(), 'viewer', 'Read-only access', ARRAY[
        'workflow:read', 'node:read', 'execution:read'
    ], TRUE),
    (gen_random_uuid(), 'executor', 'Can execute workflows', ARRAY[
        'workflow:read', 'workflow:execute', 'execution:read'
    ], TRUE)
ON CONFLICT (name) DO NOTHING;

-- Add comment
COMMENT ON TABLE roles IS 'User roles with associated permissions';
COMMENT ON TABLE user_roles IS 'Many-to-many relationship between users and roles';
