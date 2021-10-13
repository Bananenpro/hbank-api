package services

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"

	"github.com/Bananenpro/hbank-api/config"
)

var emailAuth smtp.Auth

func EmailAuthenticate() {
	emailAuth = smtp.PlainAuth("", config.Data.EmailUsername, config.Data.EmailPassword, config.Data.EmailHost)
}

func ParseEmailTemplate(name string, lang string, data interface{}) (string, error) {
	filepath := fmt.Sprintf("templates/email/%s/%s.html", lang, name)

	t, err := template.ParseFiles(filepath)
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)
	if err = t.Execute(buf, data); err != nil {
		return "", err
	}

	body := buf.String()
	return body, nil
}

func SendEmail(address []string, subject string, body string) error {
	if !config.Data.EmailEnabled {
		return nil
	}
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"

	subject = "Subject: " + subject + "\n"
	msg := []byte(subject + mime + "\n" + body)
	addr := fmt.Sprintf("%s:%d", config.Data.EmailHost, config.Data.EmailPort)

	return smtp.SendMail(addr, emailAuth, config.Data.EmailUsername, address, msg)
}
