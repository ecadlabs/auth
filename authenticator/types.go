package authenticator

import (
	"context"
)

// Credentials represents name/secret pair
type Credentials struct {
	ID     string // email etc
	Secret []byte // password etc
}

// Error extends error interface with Rejected() method
type Error interface {
	error
	Rejected() bool // Is request rejected?
}

// Authenticator interface
type Authenticator interface {
	Authenticate(context.Context, *Credentials) (Result, error)
}

// Result contains backend data
type Result interface {
	Claims() map[string]interface{}
}
