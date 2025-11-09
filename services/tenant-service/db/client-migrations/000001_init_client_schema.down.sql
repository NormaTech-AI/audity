-- Rollback client-specific database schema

-- Drop triggers
DROP TRIGGER IF EXISTS update_reports_updated_at ON reports;
DROP TRIGGER IF EXISTS update_comments_updated_at ON comments;
DROP TRIGGER IF EXISTS update_submissions_updated_at ON submissions;
DROP TRIGGER IF EXISTS update_questions_updated_at ON questions;
DROP TRIGGER IF EXISTS update_audits_updated_at ON audits;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables (in reverse dependency order)
DROP TABLE IF EXISTS activity_log;
DROP TABLE IF EXISTS reports;
DROP TABLE IF EXISTS comments;
DROP TABLE IF EXISTS evidence;
DROP TABLE IF EXISTS submissions;
DROP TABLE IF EXISTS question_assignments;
DROP TABLE IF EXISTS questions;
DROP TABLE IF EXISTS audits;

-- Drop enums
DROP TYPE IF EXISTS report_status_enum;
DROP TYPE IF EXISTS audit_status_enum;
DROP TYPE IF EXISTS submission_status_enum;
DROP TYPE IF EXISTS answer_value_enum;
DROP TYPE IF EXISTS question_type_enum;
