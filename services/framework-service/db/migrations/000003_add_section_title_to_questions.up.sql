-- Add section_title column to framework_questions
ALTER TABLE framework_questions ADD COLUMN section_title TEXT;

-- Create index for better query performance
CREATE INDEX idx_framework_questions_section_title ON framework_questions(section_title);
