package notification

import (
	"git.ecadlabs.com/ecad/auth/users"
	log "github.com/sirupsen/logrus"
	"time"
)

type Data struct {
	Self   *users.User
	User   *users.User
	Token  string
	MaxAge time.Duration
}

type Notifier interface {
	NewUser(*Data) error
	PasswordReset(*Data) error
}

// For debug purpose
type Log struct{}

func (l Log) NewUser(d *Data) error {
	log.WithFields(log.Fields{
		"id":    d.User.ID,
		"token": d.Token,
	}).Println("Reset token")

	return nil
}

func (l Log) PasswordReset(d *Data) error {
	log.WithFields(log.Fields{
		"id":    d.User.ID,
		"email": d.User.Email,
		"token": d.Token,
	}).Println("Reset token requested")

	return nil
}
