-- Create audit cycles table
CREATE TABLE audit_cycles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    status VARCHAR(50) DEFAULT 'active' CHECK (status IN ('active', 'completed', 'archived')),
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT valid_date_range CHECK (end_date >= start_date)
);

CREATE INDEX idx_audit_cycles_status ON audit_cycles(status);
CREATE INDEX idx_audit_cycles_dates ON audit_cycles(start_date, end_date);

-- Create audit cycle clients table (many-to-many relationship)
CREATE TABLE audit_cycle_clients (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    audit_cycle_id UUID NOT NULL REFERENCES audit_cycles(id) ON DELETE CASCADE,
    client_id UUID NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(audit_cycle_id, client_id)
);

CREATE INDEX idx_audit_cycle_clients_cycle_id ON audit_cycle_clients(audit_cycle_id);
CREATE INDEX idx_audit_cycle_clients_client_id ON audit_cycle_clients(client_id);

-- Create audit cycle frameworks table (frameworks assigned to clients in a cycle)
CREATE TABLE audit_cycle_frameworks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    audit_cycle_client_id UUID NOT NULL REFERENCES audit_cycle_clients(id) ON DELETE CASCADE,
    framework_id UUID NOT NULL, -- References framework in framework-service
    framework_name VARCHAR(255) NOT NULL,
    assigned_by UUID REFERENCES users(id) ON DELETE SET NULL,
    assigned_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    due_date DATE,
    status VARCHAR(50) DEFAULT 'pending' CHECK (status IN ('pending', 'in_progress', 'completed', 'overdue')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_audit_cycle_frameworks_cycle_client_id ON audit_cycle_frameworks(audit_cycle_client_id);
CREATE INDEX idx_audit_cycle_frameworks_status ON audit_cycle_frameworks(status);

-- Triggers for updated_at
CREATE TRIGGER update_audit_cycles_updated_at BEFORE UPDATE ON audit_cycles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_audit_cycle_frameworks_updated_at BEFORE UPDATE ON audit_cycle_frameworks
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
