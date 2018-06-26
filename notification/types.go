package notification

import (
	"context"
	"time"

	"git.ecadlabs.com/ecad/auth/users"
	log "github.com/sirupsen/logrus"
)

type NotificationData struct {
	CurrentUser *users.User
	TargetUser  *users.User
	Token       string
	TokenMaxAge time.Duration
}

type Notifier interface {
	InviteUser(context.Context, *NotificationData) error
	PasswordReset(context.Context, *NotificationData) error
}

// For debug purpose
type Log struct{}

func (l Log) InviteUser(ctx context.Context, d *NotificationData) error {
	log.WithFields(log.Fields{
		"id":    d.TargetUser.ID,
		"token": d.Token,
	}).Println("Reset token")

	return nil
}

func (l Log) PasswordReset(ctx context.Context, d *NotificationData) error {
	log.WithFields(log.Fields{
		"id":    d.TargetUser.ID,
		"email": d.TargetUser.Email,
		"token": d.Token,
	}).Println("Reset token requested")

	return nil
}

var _ Notifier = Log{}
