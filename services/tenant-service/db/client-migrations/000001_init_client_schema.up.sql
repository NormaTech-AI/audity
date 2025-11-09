-- Client-Specific Database Schema
-- This schema is deployed to each tenant's isolated PostgreSQL database
-- It contains audit-specific data: questions, evidence, submissions, etc.

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================
-- ENUMS
-- ============================================

-- Question types (Yes/No, Text, etc.)
CREATE TYPE question_type_enum AS ENUM (
    'yes_no',
    'text',
    'multiple_choice'
);

-- Answer values for yes/no questions
CREATE TYPE answer_value_enum AS ENUM (
    'yes',
    'no',
    'na'
);

-- Submission status for each question
CREATE TYPE submission_status_enum AS ENUM (
    'not_started',
    'in_progress',
    'submitted',
    'approved',
    'rejected',
    'referred'
);

-- Overall audit status
CREATE TYPE audit_status_enum AS ENUM (
    'not_started',
    'in_progress',
    'under_review',
    'completed',
    'overdue'
);

-- Report status
CREATE TYPE report_status_enum AS ENUM (
    'pending',
    'generated',
    'signed',
    'delivered'
);

-- ============================================
-- TABLES
-- ============================================

-- Audit Assignments
-- Links a framework to this client with due date
CREATE TABLE audits (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    framework_id UUID NOT NULL, -- References compliance_frameworks in tenant_db
    framework_name VARCHAR(100) NOT NULL, -- Denormalized for quick access
    assigned_by UUID NOT NULL, -- User ID from tenant_db
    assigned_to UUID, -- Client POC user ID
    due_date DATE NOT NULL,
    status audit_status_enum NOT NULL DEFAULT 'not_started',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE
);

-- Questions from compliance frameworks
-- Master list of questions for this tenant's assigned frameworks
CREATE TABLE questions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    audit_id UUID NOT NULL REFERENCES audits(id) ON DELETE CASCADE,
    section VARCHAR(255) NOT NULL, -- e.g., "Network Security", "Access Control"
    question_number VARCHAR(50) NOT NULL, -- e.g., "1.1", "2.3.4"
    question_text TEXT NOT NULL,
    question_type question_type_enum NOT NULL DEFAULT 'yes_no',
    help_text TEXT, -- Additional guidance for the question
    is_mandatory BOOLEAN NOT NULL DEFAULT true,
    display_order INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(audit_id, question_number)
);

-- Question Assignments
-- Delegation of questions to specific stakeholders
CREATE TABLE question_assignments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    question_id UUID NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    assigned_to UUID NOT NULL, -- User ID of stakeholder
    assigned_by UUID NOT NULL, -- User ID of POC who delegated
    assigned_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    notes TEXT,
    UNIQUE(question_id, assigned_to)
);

-- Question Submissions
-- Answers and evidence submitted by client
CREATE TABLE submissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    question_id UUID NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    submitted_by UUID NOT NULL, -- User ID who submitted
    answer_value answer_value_enum, -- For yes/no questions
    answer_text TEXT, -- For text questions or explanation
    explanation TEXT NOT NULL, -- Mandatory explanation for all answers
    status submission_status_enum NOT NULL DEFAULT 'not_started',
    submitted_at TIMESTAMP WITH TIME ZONE,
    reviewed_by UUID, -- Auditor who reviewed
    reviewed_at TIMESTAMP WITH TIME ZONE,
    review_notes TEXT, -- Auditor's comments
    rejection_reason TEXT, -- Reason if rejected
    version INTEGER NOT NULL DEFAULT 1, -- Incremented on resubmission
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Evidence Files
-- Files uploaded as evidence for submissions
CREATE TABLE evidence (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    submission_id UUID NOT NULL REFERENCES submissions(id) ON DELETE CASCADE,
    file_name VARCHAR(255) NOT NULL,
    file_path VARCHAR(500) NOT NULL, -- MinIO object path
    file_size BIGINT NOT NULL, -- Size in bytes
    file_type VARCHAR(100), -- MIME type
    uploaded_by UUID NOT NULL,
    uploaded_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    description TEXT,
    is_deleted BOOLEAN NOT NULL DEFAULT false,
    deleted_at TIMESTAMP WITH TIME ZONE,
    deleted_by UUID
);

-- Comments/Discussion
-- Internal and client comments on submissions
CREATE TABLE comments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    submission_id UUID NOT NULL REFERENCES submissions(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    user_name VARCHAR(255) NOT NULL, -- Denormalized
    comment_text TEXT NOT NULL,
    is_internal BOOLEAN NOT NULL DEFAULT false, -- Internal Nishaj comments
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Audit Reports
-- Final generated reports
CREATE TABLE reports (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    audit_id UUID NOT NULL REFERENCES audits(id) ON DELETE CASCADE,
    unsigned_file_path VARCHAR(500), -- MinIO path to unsigned PDF
    signed_file_path VARCHAR(500), -- MinIO path to signed PDF
    generated_by UUID NOT NULL, -- Auditor who generated
    generated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    signed_by UUID, -- Auditor who uploaded signed version
    signed_at TIMESTAMP WITH TIME ZONE,
    status report_status_enum NOT NULL DEFAULT 'pending',
    metadata JSONB, -- Additional report metadata
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(audit_id) -- One report per audit
);

-- Activity Log
-- Audit trail for all actions in this tenant's database
CREATE TABLE activity_log (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    user_email VARCHAR(255) NOT NULL,
    action VARCHAR(100) NOT NULL, -- e.g., "submission_created", "evidence_uploaded"
    entity_type VARCHAR(50) NOT NULL, -- e.g., "submission", "question", "report"
    entity_id UUID NOT NULL,
    details JSONB, -- Additional context
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- ============================================
-- INDEXES
-- ============================================

-- Audits
CREATE INDEX idx_audits_status ON audits(status);
CREATE INDEX idx_audits_due_date ON audits(due_date);

-- Questions
CREATE INDEX idx_questions_audit_id ON questions(audit_id);
CREATE INDEX idx_questions_section ON questions(section);
CREATE INDEX idx_questions_display_order ON questions(audit_id, display_order);

-- Question Assignments
CREATE INDEX idx_question_assignments_assigned_to ON question_assignments(assigned_to);
CREATE INDEX idx_question_assignments_question_id ON question_assignments(question_id);

-- Submissions
CREATE INDEX idx_submissions_question_id ON submissions(question_id);
CREATE INDEX idx_submissions_submitted_by ON submissions(submitted_by);
CREATE INDEX idx_submissions_status ON submissions(status);
CREATE INDEX idx_submissions_reviewed_by ON submissions(reviewed_by);

-- Evidence
CREATE INDEX idx_evidence_submission_id ON evidence(submission_id);
CREATE INDEX idx_evidence_uploaded_by ON evidence(uploaded_by);
CREATE INDEX idx_evidence_uploaded_at ON evidence(uploaded_at DESC);

-- Comments
CREATE INDEX idx_comments_submission_id ON comments(submission_id);
CREATE INDEX idx_comments_created_at ON comments(created_at DESC);

-- Reports
CREATE INDEX idx_reports_audit_id ON reports(audit_id);
CREATE INDEX idx_reports_status ON reports(status);

-- Activity Log
CREATE INDEX idx_activity_log_user_id ON activity_log(user_id);
CREATE INDEX idx_activity_log_entity ON activity_log(entity_type, entity_id);
CREATE INDEX idx_activity_log_created_at ON activity_log(created_at DESC);

-- ============================================
-- TRIGGERS
-- ============================================

-- Auto-update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_audits_updated_at BEFORE UPDATE ON audits
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_questions_updated_at BEFORE UPDATE ON questions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_submissions_updated_at BEFORE UPDATE ON submissions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_comments_updated_at BEFORE UPDATE ON comments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_reports_updated_at BEFORE UPDATE ON reports
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- COMMENTS
-- ============================================

COMMENT ON TABLE audits IS 'Audit assignments for this tenant';
COMMENT ON TABLE questions IS 'Questions from compliance frameworks';
COMMENT ON TABLE question_assignments IS 'Delegation of questions to stakeholders';
COMMENT ON TABLE submissions IS 'Client answers and submissions';
COMMENT ON TABLE evidence IS 'Evidence files uploaded by client';
COMMENT ON TABLE comments IS 'Discussion and comments on submissions';
COMMENT ON TABLE reports IS 'Generated audit reports';
COMMENT ON TABLE activity_log IS 'Audit trail of all tenant activities';
