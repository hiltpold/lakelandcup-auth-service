package utils

import (
	"os"

	"github.com/sendgrid/rest"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

func SendGridMail(name string, email string, subject string, fileName string, token string, sgKey string) (*rest.Response, error) {
	from := mail.NewEmail("Lakelandcup", os.Getenv("SENDGRID_EMAIL"))
	to := mail.NewEmail(name, email)
	subjectMail := subject
	template := ParseHtml(fileName, map[string]string{
		"to":            email,
		"token":         token,
		"activationUrl": os.Getenv("ACTIVATION_URL"),
	})
	message := mail.NewSingleEmail(from, subjectMail, to, "", template)
	client := sendgrid.NewSendClient(sgKey)
	return client.Send(message)
}
