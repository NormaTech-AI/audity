-- Drop framework_questions table
DROP TRIGGER IF EXISTS update_framework_questions_updated_at ON framework_questions;
DROP INDEX IF EXISTS idx_framework_questions_control_id;
DROP INDEX IF EXISTS idx_framework_questions_framework_id;
DROP TABLE IF EXISTS framework_questions;

-- Restore checklist_json column if it was dropped
-- ALTER TABLE compliance_frameworks ADD COLUMN checklist_json JSONB NOT NULL DEFAULT '{}'::jsonb;
