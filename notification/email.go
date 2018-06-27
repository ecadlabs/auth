package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/mail"

	log "github.com/sirupsen/logrus"
)

type Message struct {
	To      *mail.Address
	From    *mail.Address
	Subject string
	Body    []byte
}

type EmailTemplateData struct {
	ResetURLPrefix string `json:"reset_url_prefix"`
	AppName        string `json:"app_name"`
}

type MailDriver interface {
	SendMessage(*Message) error
}

type EmailNotifier struct {
	ch     chan *Message
	driver MailDriver
	from   *mail.Address
	data   *EmailTemplateData
}

type DriverFunc func(json.RawMessage) (MailDriver, error)

var driverRegistry = make(map[string]DriverFunc)

func RegisterDriver(name string, f DriverFunc) {
	driverRegistry[name] = f
}

func NewEmailNotifier(from *mail.Address, templateData *EmailTemplateData, driver string, data json.RawMessage) (*EmailNotifier, error) {
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
		data:   templateData,
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

func (e *EmailNotifier) send(ctx context.Context, d *NotificationData, tplPrefix string) error {
	templateData := struct {
		*EmailTemplateData
		*NotificationData
	}{
		NotificationData:  d,
		EmailTemplateData: e.data,
	}

	var subject bytes.Buffer
	if err := emailTemplates.ExecuteTemplate(&subject, tplPrefix+"_subject", &templateData); err != nil {
		return err
	}

	var body bytes.Buffer
	if err := emailTemplates.ExecuteTemplate(&body, tplPrefix+"_body", &templateData); err != nil {
		return err
	}

	msg := Message{
		To:      &mail.Address{Name: d.TargetUser.Name, Address: d.TargetUser.Email},
		From:    e.from,
		Subject: subject.String(),
		Body:    body.Bytes(),
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case e.ch <- &msg:
	}

	return nil
}

func (e *EmailNotifier) InviteUser(ctx context.Context, d *NotificationData) error {
	return e.send(ctx, d, "invite")
}

func (e *EmailNotifier) PasswordReset(ctx context.Context, d *NotificationData) error {
	return e.send(ctx, d, "reset")
}

var _ Notifier = &EmailNotifier{}
