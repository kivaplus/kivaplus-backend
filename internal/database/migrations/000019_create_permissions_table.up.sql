-- Create permissions table for granular, custom role management
-- This allows for flexible, custom roles while keeping predefined roles in JSONB

CREATE TABLE permissions (
    id BIGSERIAL PRIMARY KEY,
    role_id INTEGER NOT NULL REFERENCES role(id) ON DELETE CASCADE,
    resource VARCHAR(50) NOT NULL,  -- e.g., 'condominiums', 'employees', 'vehicles'
    action VARCHAR(20) NOT NULL,    -- e.g., 'create', 'read', 'update', 'delete'
    allowed BOOLEAN NOT NULL DEFAULT true,
    scope VARCHAR(20) DEFAULT 'global', -- 'global', 'condominium', 'own'
    condominium_id BIGINT REFERENCES condominium(id) ON DELETE CASCADE, -- For condominium-scoped permissions
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Ensure unique permission per role/resource/action/scope combination
    CONSTRAINT uk_role_resource_action_scope UNIQUE(role_id, resource, action, scope, condominium_id)
);

-- Indexes for performance
CREATE INDEX idx_permissions_role_id ON permissions(role_id);
CREATE INDEX idx_permissions_resource ON permissions(resource);
CREATE INDEX idx_permissions_action ON permissions(action);
CREATE INDEX idx_permissions_scope ON permissions(scope);
CREATE INDEX idx_permissions_condominium_id ON permissions(condominium_id);

-- Add trigger for updated_at
CREATE TRIGGER update_permissions_updated_at
    BEFORE UPDATE ON permissions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Add a column to role table to indicate if it uses JSONB or granular permissions
ALTER TABLE role ADD COLUMN permission_type VARCHAR(20) NOT NULL DEFAULT 'jsonb' CHECK (permission_type IN ('jsonb', 'granular'));

-- Update existing roles to use JSONB type
UPDATE role SET permission_type = 'jsonb';

-- Example: Create a custom role that uses granular permissions
INSERT INTO role (name, description, permission_type, permissions) VALUES
('custom_manager', 'Custom Manager Role - Uses granular permissions', 'granular', '{}');

-- Add granular permissions for the custom role
INSERT INTO permissions (role_id, resource, action, allowed, scope) VALUES
-- Custom manager can read all condominiums globally
((SELECT id FROM role WHERE name = 'custom_manager'), 'condominiums', 'read', true, 'global'),
-- But can only update condominiums they manage
((SELECT id FROM role WHERE name = 'custom_manager'), 'condominiums', 'update', true, 'condominium'),
-- Can manage employees in their condominiums
((SELECT id FROM role WHERE name = 'custom_manager'), 'employees', 'create', true, 'condominium'),
((SELECT id FROM role WHERE name = 'custom_manager'), 'employees', 'read', true, 'condominium'),
((SELECT id FROM role WHERE name = 'custom_manager'), 'employees', 'update', true, 'condominium'),
-- Can read vehicles globally but only manage in their condominium
((SELECT id FROM role WHERE name = 'custom_manager'), 'vehicles', 'read', true, 'global'),
((SELECT id FROM role WHERE name = 'custom_manager'), 'vehicles', 'create', true, 'condominium'),
((SELECT id FROM role WHERE name = 'custom_manager'), 'vehicles', 'update', true, 'condominium');
