package notification

import (
	"bytes"
	"fmt"
	"net"
	"net/smtp"
	"strings"

	"git.ecadlabs.com/ecad/auth/utils"
)

type SMTPDriver struct {
	Address  string
	Username string
	Password string
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

	toHeaderList := make([]string, len(msg.To))
	toList := make([]string, len(msg.To))

	for i, a := range msg.To {
		toHeaderList[i] = a.String()
		toList[i] = a.Address
	}

	var body bytes.Buffer

	fmt.Fprintf(&body, "From: %s\n", msg.From.String())
	fmt.Fprintf(&body, "To: %s\n", strings.Join(toHeaderList, ", "))
	fmt.Fprintf(&body, "Subject: %s\n", msg.Subject)
	body.WriteString("Mime-Version: 1.0\nContent-Type: text/plain; charset=UTF-8\n\n")
	body.Write(msg.Body)

	return smtp.SendMail(s.Address, auth, msg.From.Address, toList, body.Bytes())
}

func newSMTPDriver(opt utils.Options) (MailDriver, error) {
	var d SMTPDriver
	d.Address, _ = opt.GetString("address")
	d.Username, _ = opt.GetString("user")
	d.Password, _ = opt.GetString("password")

	return &d, nil
}

func init() {
	RegisterDriver("smtp", newSMTPDriver)
}
