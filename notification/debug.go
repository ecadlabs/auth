package notification

import (
	"encoding/json"
	"fmt"
)

type debugDriver struct{}

func (d debugDriver) SendMessage(msg *Message) error {
	fmt.Printf("From: %s\nTo: %s\nSubject: %s\n", msg.From.String(), msg.To.String(), msg.Subject)
	fmt.Println("Body:")
	fmt.Println(string(msg.Body))
	return nil
}

func newDebugDriver(data json.RawMessage) (MailDriver, error) {
	return debugDriver{}, nil
}

func init() {
	RegisterDriver("debug", newDebugDriver)
}
