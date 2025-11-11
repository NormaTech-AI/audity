package mail

import (
	"github.com/NormaTech-AI/audity/packages/go/microsoft-mail"
	"github.com/NormaTech-AI/audity/services/tenant-service/internal/config"
	"go.uber.org/zap"
)

type MailService struct {
	client *microsoftmail.Client
}

var Mail *MailService

func NewMailService(cfg *config.Config, log *zap.SugaredLogger) (*MailService, error) {
	client, err := microsoftmail.NewClient(microsoftmail.Config{
		TenantID:     cfg.MicrosoftMail.TenantID,
		ClientID:     cfg.MicrosoftMail.ClientID,
		ClientSecret: cfg.MicrosoftMail.ClientSecret,
		SenderEmail:  cfg.MicrosoftMail.SenderEmail,
	}, log)
	if err != nil {
		log.Errorw("failed to create mail client", "error", err)
		return nil, err
	}
	log.Info("mail client created successfully")
	Mail = &MailService{client: client}
	return Mail, nil
}
// SendEmail sends an email using the configured mail client
func (m *MailService) SendEmail(to []string, subject, body string, log *zap.SugaredLogger) error {
	return m.client.SendEmail(microsoftmail.EmailParams{
		To:          to,
		Subject:     subject,
		Body:        body,
		ContentType: "HTML",
	}, log)
}

// SendMail sends an email using the configured mail client
func (m *MailService) SendMail(to []string, subject, body string, log *zap.SugaredLogger) error {
	return m.client.SendEmail(microsoftmail.EmailParams{
		To:          to,
		Subject:     subject,
		Body:        body,
		ContentType: "HTML",
	}, log)
}
