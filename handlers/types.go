package handlers

import (
	"github.com/ecadlabs/auth/storage"
)

type Storage interface {
	storage.UserStorage
	storage.MembershipStorage
	storage.TenantStorage
	storage.LogStorage
}
