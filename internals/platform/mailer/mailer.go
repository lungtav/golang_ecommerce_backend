package mailer

import (
	"log/slog"

	"github.com/lungtav/golang_ecommerce_backend/internals/platform/mailtemplate"
	"github.com/resend/resend-go/v4"
)

type Mailer struct {
	client *resend.Client
}

func New(key string) *Mailer {
	return &Mailer{client: resend.NewClient(key)}
}

func (m *Mailer) SendEmailVerification(name, email, url string) (*resend.SendEmailResponse, error) {
	emailVerficationParams := &resend.SendEmailRequest{
		From:    "Ecommerce in Go <verify@oluwafunmbi.cv>",
		To:      []string{email},
		Html:    mailtemplate.VerifyEmail(name, url),
		Subject: "Verify your email",
	}

	sent, err := m.client.Emails.Send(emailVerficationParams)

	if err != nil {
		slog.Error("verifcation email failed to send", "error", err)
		return sent, err
	}

	return sent, nil
}
