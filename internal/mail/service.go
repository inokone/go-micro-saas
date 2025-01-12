package mail

import (
	"bytes"
	_ "embed"
	"errors"
	"html/template"
	"time"

	"github.com/cskr/pubsub/v2"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"gopkg.in/mail.v2"

	"github.com/inokone/go-micro-saas/internal/common"
)

const (
	confirmation = "confirmation"
	pwdReset     = "passwordreset"
)

//go:embed "confirmation.html"
var ct string

//go:embed "passwordreset.html"
var pt string

// Service is a struct for a service sending mails for our users.
type Service struct {
	config    *common.MailConfig
	dialer    *mail.Dialer
	ps        *pubsub.PubSub[string, common.Event]
	templates map[string]*template.Template
}

type SendRequest struct {
	UserID    uuid.UUID
	Recipient string
	Subject   string
	Template  string
	Data      interface{}
	App       string
}

// NewService create a new `Service` entity based on the configuration.
// If SMTP server is not configured the service will not return error, just logs it as a warning.
func NewService(config *common.MailConfig, ps *pubsub.PubSub[string, common.Event]) *Service {
	if len(config.SMTPAddress) == 0 {
		log.Warn("SMTP is not set up, e-mail sending functionality will not work correctly!")
	}

	return &Service{
		config:    config,
		dialer:    mail.NewDialer(config.SMTPAddress, config.SMTPPort, config.SMTPUser, config.SMTPPassword),
		templates: loadTemplates(),
		ps:        ps,
	}
}

func loadTemplates() map[string]*template.Template {
	return map[string]*template.Template{
		confirmation: mustLoadTemplate(ct),
		pwdReset:     mustLoadTemplate(pt),
	}
}

func mustLoadTemplate(tpl string) *template.Template {
	tmpl, err := template.New("email").Parse(tpl)
	if err != nil {
		log.WithError(err).Error("email template can not be parsed")
		panic("email template can not be parsed")
	}
	return tmpl
}

type templateData struct {
	Link string
	App  string
}

// Send is a method of `Service` sends an e-mail to the recipient email address with the subject and body provided as parameters
// If SMTP server is not configured the service will not return error, just logs it as a warning.
func (s *Service) send(recipient string, subject string, body string, userID uuid.UUID) error {
	if len(s.config.SMTPAddress) == 0 {
		log.Warn("SMTP is not set up, failed to send the e-mail!")
		return nil
	}
	// Set up email message
	m := mail.NewMessage()
	m.SetHeader("From", s.config.NoReplyAddress)
	m.SetHeader("To", recipient)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	// Send the email
	err := s.dialer.DialAndSend(m)

	if err == nil {
		s.ps.Pub(common.Event{
			Type: common.EmailSent,
			Time: time.Now(),
			User: userID,
			ID:   uuid.New(),
			Data: common.EmailData{
				From:    s.config.NoReplyAddress,
				To:      recipient,
				Subject: subject,
				Body:    body,
			},
		}, common.HistoryTopic)
	}

	return err
}

func (s *Service) Send(r *SendRequest) error {
	var (
		c bytes.Buffer
		t *template.Template
	)
	t = s.templates[r.Template]
	if t == nil {
		return errors.New(r.Template + " template not found")
	}
	if err := t.Execute(&c, r.Data); err != nil {
		return err
	}
	return s.send(r.Recipient, r.Subject, c.String(), r.UserID)
}

// EmailConfirmation is a method of `Service` sends an e-mail confirmation message to the recipient email address
func (s *Service) EmailConfirmation(recipient string, confirmationURL string) error {
	return s.Send(&SendRequest{
		UserID:    uuid.Nil,
		Recipient: recipient,
		Subject:   "E-mail Confirmation",
		Template:  confirmation,
		Data: templateData{
			Link: confirmationURL,
			App:  s.config.ApplicationName,
		},
	})
}

// PasswordReset is a method of `Service` sends a password reset message to the recipient email address
func (s *Service) PasswordReset(recipient string, resetURL string) error {
	return s.Send(&SendRequest{
		UserID:    uuid.Nil,
		Recipient: recipient,
		Subject:   "Password Reset",
		Template:  pwdReset,
		Data: templateData{
			Link: resetURL,
			App:  s.config.ApplicationName,
		},
	})
}
