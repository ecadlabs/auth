package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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

type MailDriver interface {
	SendMessage(*Message) error
}

type EmailNotifier struct {
	ch     chan *Message
	driver MailDriver
	from   *mail.Address
}

type DriverFunc func(json.RawMessage) (MailDriver, error)

var driverRegistry = make(map[string]DriverFunc)

func RegisterDriver(name string, f DriverFunc) {
	driverRegistry[name] = f
}

func NewEmailNotifier(from *mail.Address, driver string, data json.RawMessage) (*EmailNotifier, error) {
	driverFunc, ok := driverRegistry[driver]
	if !ok {
		return nil, fmt.Errorf("Unknown driver `%s'", driver)
	}

	drv, err := driverFunc(data)
	if err != nil {
		return nil, err
	}

	n := EmailNotifier{
		ch:     make(chan *Message, 100),
		driver: drv,
		from:   from,
	}

	// Send in background
	go func() {
		for msg := range n.ch {
			if err := n.driver.SendMessage(msg); err != nil {
				log.Error(err)
			}
		}
	}()

	return &n, nil
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
