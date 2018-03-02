package mem

import (
	"context"
	"git.ecadlabs.com/ecad/auth/authenticator"
	"golang.org/x/crypto/bcrypt"
)

type memError string

func (m memError) Rejected() bool {
	return true
}

func (m memError) Error() string {
	return string(m)
}

// User represents basic user data
type User struct {
	Name         string
	PasswordHash []byte
	Email        string
}

// UsersMap is simple in-memory user database. For debugging purposes only
type UsersMap map[string]User

// Authenticate performs name lookup and password hash comparison
func (u UsersMap) Authenticate(ctx context.Context, cred *authenticator.Credentials) (map[string]interface{}, error) {
	if user, ok := u[cred.ID]; ok {
		if bcrypt.CompareHashAndPassword(user.PasswordHash, cred.Secret) == nil {
			return nil, nil
		}
	}

	return nil, memError("Forbidden")
}

var _ authenticator.Authenticator = UsersMap{}
