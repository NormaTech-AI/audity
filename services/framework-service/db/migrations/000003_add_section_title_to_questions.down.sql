-- Remove section_title column from framework_questions
DROP INDEX IF EXISTS idx_framework_questions_section_title;
ALTER TABLE framework_questions DROP COLUMN IF EXISTS section_title;
