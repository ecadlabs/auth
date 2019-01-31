package notification

import (
	"bytes"
	"context"
	"fmt"
	"net/mail"
	"time"

	"github.com/ecadlabs/auth/utils"
	log "github.com/sirupsen/logrus"
)

type Message struct {
	To      []*mail.Address
	From    *mail.Address
	Subject string
	Body    []byte
}

type EmailTemplateData struct {
	TenantInvitePrefix   string `json:"tenant_invite_prefix"`
	ResetURLPrefix       string `json:"reset_url_prefix"`
	UpdateEmailURLPrefix string `json:"update_email_prefix"`
	AppName              string `json:"app_name"`
	SupportEmail         string `json:"support_email"`
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

type DriverFunc func(utils.Options) (MailDriver, error)

var driverRegistry = make(map[string]DriverFunc)

func RegisterDriver(name string, f DriverFunc) {
	driverRegistry[name] = f
}

func NewEmailNotifier(from *mail.Address, templateData *EmailTemplateData, driver string, data utils.Options) (*EmailNotifier, error) {
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
		Timestamp time.Time `json:"timestamp"`
	}{
		NotificationData:  d,
		EmailTemplateData: e.data,
		Timestamp:         time.Now(),
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
		From:    e.from,
		Subject: subject.String(),
		Body:    body.Bytes(),
	}

	if len(d.To) != 0 {
		msg.To = make([]*mail.Address, len(d.To))

		for i, addr := range d.To {
			msg.To[i] = &mail.Address{Address: addr}
		}
	} else {
		msg.To = []*mail.Address{&mail.Address{Name: d.TargetUser.Name, Address: d.TargetUser.Email}}
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case e.ch <- &msg:
	}

	return nil
}

func (e *EmailNotifier) Notify(ctx context.Context, tpl string, d *NotificationData) error {
	return e.send(ctx, d, tpl)
}

var _ Notifier = &EmailNotifier{}
