package notification

import (
	"bytes"
	"context"
	log "github.com/sirupsen/logrus"
	"net/mail"
	"text/template"
)

type Message struct {
	To      *mail.Address
	From    *mail.Address
	Subject string
	Body    []byte
}

type MessageSender interface {
	SendMessage(*Message) error
}

type EmailNotifier struct {
	ch     chan *Message
	sender MessageSender
	from   *mail.Address
}

func NewEmailNotifier(from *mail.Address, backend MessageSender) *EmailNotifier {
	n := EmailNotifier{
		ch:     make(chan *Message, 100),
		sender: backend,
		from:   from,
	}

	// Send in background
	go func() {
		for msg := range n.ch {
			if err := n.sender.SendMessage(msg); err != nil {
				log.Error(err)
			}
		}
	}()

	return &n
}

// TODO
var newUserTemplate = template.Must(template.New("message").Parse("Token: {{.Token}}\r\nExpires after: {{.TokenMaxAge}}\r\n"))

func (e *EmailNotifier) NewUser(ctx context.Context, d *ResetData) error {
	var body bytes.Buffer

	if err := newUserTemplate.Execute(&body, d); err != nil {
		return err
	}

	msg := Message{
		To:      &mail.Address{Name: d.TargetUser.Name, Address: d.TargetUser.Email},
		From:    e.from,
		Subject: "New user", // TODO
		Body:    body.Bytes(),
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case e.ch <- &msg:
	}

	return nil
}

func (e *EmailNotifier) PasswordReset(ctx context.Context, d *ResetData) error {
	return e.NewUser(ctx, d) // TODO
}

var _ Notifier = &EmailNotifier{}
