package framework

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/NormaTech-AI/audity/services/tenant-service/internal/clientdb"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
)

// FrameworkTemplate represents a compliance framework template
type FrameworkTemplate struct {
	FrameworkName string    `json:"framework_name"`
	Version       string    `json:"version"`
	Sections      []Section `json:"sections"`
}

// Section represents a section in a framework
type Section struct {
	Name      string     `json:"name"`
	Questions []Question `json:"questions"`
}

// Question represents a question in a framework
type Question struct {
	Number      string `json:"number"`
	Text        string `json:"text"`
	Type        string `json:"type"`
	HelpText    string `json:"help_text"`
	IsMandatory bool   `json:"is_mandatory"`
}

// Service handles framework template operations
type Service struct {
	templatesDir string
	logger       *zap.SugaredLogger
}

// NewService creates a new framework service
func NewService(templatesDir string, logger *zap.SugaredLogger) *Service {
	return &Service{
		templatesDir: templatesDir,
		logger:       logger,
	}
}

// LoadTemplate loads a framework template from a JSON file
func (s *Service) LoadTemplate(frameworkName string) (*FrameworkTemplate, error) {
	filename := filepath.Join(s.templatesDir, fmt.Sprintf("%s-template.json", frameworkName))
	
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read template file: %w", err)
	}

	var template FrameworkTemplate
	if err := json.Unmarshal(data, &template); err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	return &template, nil
}

// PopulateQuestions creates questions in a client database from a framework template
func (s *Service) PopulateQuestions(ctx context.Context, queries *clientdb.Queries, auditID uuid.UUID, frameworkName string) error {
	s.logger.Infow("Populating questions from template", "framework", frameworkName, "audit_id", auditID)

	// Load template
	template, err := s.LoadTemplate(frameworkName)
	if err != nil {
		return fmt.Errorf("failed to load template: %w", err)
	}

	displayOrder := 1
	questionCount := 0

	// Iterate through sections and create questions
	for _, section := range template.Sections {
		for _, q := range section.Questions {
			// Convert question type
			var qType clientdb.QuestionTypeEnum
			switch q.Type {
			case "yes_no":
				qType = clientdb.QuestionTypeEnumYesNo
			case "text":
				qType = clientdb.QuestionTypeEnumText
			case "multiple_choice":
				qType = clientdb.QuestionTypeEnumMultipleChoice
			default:
				qType = clientdb.QuestionTypeEnumYesNo
			}

			// Convert help text to *string
			var helpText *string
			if q.HelpText != "" {
				helpText = &q.HelpText
			}

			// Create question
			_, err := queries.CreateQuestion(ctx, clientdb.CreateQuestionParams{
				AuditID:        auditID,
				Section:        section.Name,
				QuestionNumber: q.Number,
				QuestionText:   q.Text,
				QuestionType:   qType,
				HelpText:       helpText,
				IsMandatory:    q.IsMandatory,
				DisplayOrder:   int32(displayOrder),
			})
			if err != nil {
				return fmt.Errorf("failed to create question %s: %w", q.Number, err)
			}

			displayOrder++
			questionCount++
		}
	}

	s.logger.Infow("Questions populated successfully", 
		"framework", frameworkName, 
		"audit_id", auditID, 
		"question_count", questionCount)

	return nil
}

// CreateAuditWithQuestions creates an audit and populates its questions
func (s *Service) CreateAuditWithQuestions(
	ctx context.Context,
	queries *clientdb.Queries,
	frameworkID uuid.UUID,
	frameworkName string,
	assignedBy uuid.UUID,
	assignedTo *uuid.UUID,
	dueDate time.Time,
) (uuid.UUID, error) {
	// Convert types for pgtype
	var assignedToPgtype pgtype.UUID
	if assignedTo != nil {
		assignedToPgtype = pgtype.UUID{
			Bytes: *assignedTo,
			Valid: true,
		}
	}

	dueDatePgtype := pgtype.Date{
		Time:  dueDate,
		Valid: true,
	}

	// Create audit
	audit, err := queries.CreateAudit(ctx, clientdb.CreateAuditParams{
		FrameworkID:   frameworkID,
		FrameworkName: frameworkName,
		AssignedBy:    assignedBy,
		AssignedTo:    assignedToPgtype,
		DueDate:       dueDatePgtype,
		Status:        clientdb.AuditStatusEnumNotStarted,
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create audit: %w", err)
	}

	// Populate questions from template
	if err := s.PopulateQuestions(ctx, queries, audit.ID, frameworkName); err != nil {
		return uuid.Nil, fmt.Errorf("failed to populate questions: %w", err)
	}

	return audit.ID, nil
}

// ListAvailableTemplates returns a list of available framework templates
func (s *Service) ListAvailableTemplates() ([]string, error) {
	files, err := os.ReadDir(s.templatesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read templates directory: %w", err)
	}

	templates := []string{}
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
			name := file.Name()
			// Remove "-template.json" suffix
			if len(name) > 14 {
				templates = append(templates, name[:len(name)-14])
			}
		}
	}

	return templates, nil
}
