package notification

import (
	"context"
	"git.ecadlabs.com/ecad/auth/users"
	log "github.com/sirupsen/logrus"
	"time"
)

type ResetData struct {
	CurrentUser *users.User
	TargetUser  *users.User
	Token       string
	TokenMaxAge time.Duration
}

type Notifier interface {
	NewUser(context.Context, *ResetData) error
	PasswordReset(context.Context, *ResetData) error
}

// For debug purpose
type Log struct{}

func (l Log) NewUser(ctx context.Context, d *ResetData) error {
	log.WithFields(log.Fields{
		"id":    d.TargetUser.ID,
		"token": d.Token,
	}).Println("Reset token")

	return nil
}

func (l Log) PasswordReset(ctx context.Context, d *ResetData) error {
	log.WithFields(log.Fields{
		"id":    d.TargetUser.ID,
		"email": d.TargetUser.Email,
		"token": d.Token,
	}).Println("Reset token requested")

	return nil
}

var _ Notifier = Log{}
