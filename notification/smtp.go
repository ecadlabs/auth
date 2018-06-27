package notification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/smtp"
)

type SMTPDriver struct {
	Address  string `json:"address"`
	Username string `json:"user"`
	Password string `json:"password"`
}

func (s *SMTPDriver) host() string {
	h, _, _ := net.SplitHostPort(s.Address)
	return h
}

func (s *SMTPDriver) SendMessage(msg *Message) error {
	var auth smtp.Auth

	if s.Username != "" && s.Password != "" {
		auth = smtp.PlainAuth("", s.Username, s.Password, s.host())
	}

	var body bytes.Buffer

	fmt.Fprintf(&body, "From: %s\n", msg.From.String())
	fmt.Fprintf(&body, "To: %s\n", msg.To.String())
	fmt.Fprintf(&body, "Subject: %s\n", msg.Subject)
	body.WriteString("Mime-Version: 1.0\nContent-Type: text/plain; charset=UTF-8\n\n")
	body.Write(msg.Body)

	return smtp.SendMail(s.Address,
		auth,
		msg.From.Address,
		[]string{msg.To.Address},
		body.Bytes())
}

func newSMTPDriver(data json.RawMessage) (MailDriver, error) {
	var d SMTPDriver
	if err := json.Unmarshal(data, &d); err != nil {
		return nil, err
	}

	return &d, nil
}

func init() {
	RegisterDriver("smtp", newSMTPDriver)
}
