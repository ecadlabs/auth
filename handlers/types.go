package handlers

import (
	"github.com/ecadlabs/auth/storage"
)

type Storage interface {
	storage.APIKeyStorage
	storage.UserStorage
	storage.MembershipStorage
	storage.TenantStorage
	storage.LogStorage
}
