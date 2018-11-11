package notification

import (
	"fmt"

	"git.ecadlabs.com/ecad/auth/utils"
)

type debugDriver struct{}

func (d debugDriver) SendMessage(msg *Message) error {
	fmt.Printf("From: %s\nTo: %v\nSubject: %s\n", msg.From.String(), msg.To, msg.Subject)
	fmt.Println("Body:")
	fmt.Println(string(msg.Body))
	return nil
}

func newDebugDriver(data utils.Options) (MailDriver, error) {
	return debugDriver{}, nil
}

func init() {
	RegisterDriver("debug", newDebugDriver)
}
