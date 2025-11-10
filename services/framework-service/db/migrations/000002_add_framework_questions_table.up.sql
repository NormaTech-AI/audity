-- Create framework_questions table
CREATE TABLE framework_questions (
    question_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    framework_id UUID NOT NULL REFERENCES compliance_frameworks(id) ON DELETE CASCADE,
    control_id TEXT NOT NULL,
    question_text TEXT NOT NULL,
    help_text TEXT,
    acceptable_evidence TEXT[],
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for better query performance
CREATE INDEX idx_framework_questions_framework_id ON framework_questions(framework_id);
CREATE INDEX idx_framework_questions_control_id ON framework_questions(control_id);

-- Trigger for updated_at
CREATE TRIGGER update_framework_questions_updated_at BEFORE UPDATE ON framework_questions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Remove checklist_json column from compliance_frameworks (if needed, keep for backward compatibility initially)
-- ALTER TABLE compliance_frameworks DROP COLUMN checklist_json;
