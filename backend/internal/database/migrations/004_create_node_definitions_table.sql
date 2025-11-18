-- Create nodes table (for node definitions, not workflow nodes)
CREATE TABLE IF NOT EXISTS node_definitions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    type VARCHAR(100) NOT NULL, -- trigger, action, condition, loop, etc.
    icon VARCHAR(100), -- icon name or path
    category VARCHAR(100), -- http, database, logic, etc.
    settings_schema JSONB, -- JSON schema for node settings
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    is_builtin BOOLEAN DEFAULT FALSE, -- whether this is a built-in node
    is_active BOOLEAN DEFAULT TRUE
);

-- Create indexes
CREATE INDEX idx_node_definitions_type ON node_definitions(type);
CREATE INDEX idx_node_definitions_category ON node_definitions(category);
CREATE INDEX idx_node_definitions_builtin ON node_definitions(is_builtin);
CREATE INDEX idx_node_definitions_active ON node_definitions(is_active);

-- Add updated_at trigger
CREATE TRIGGER update_node_definitions_updated_at 
    BEFORE UPDATE ON node_definitions 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();