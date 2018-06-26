package handlers

import (
	"git.ecadlabs.com/ecad/auth/logger"
	"github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
)

const (
	EvCreate       = "create"
	EvUpdate       = "update"
	EvAddRole      = "add_role"
	EvRemoveRole   = "remove_role"
	EvDelete       = "delete"
	EvReset        = "reset"
	EvResetRequest = "reset_request"
	EvLogin        = "login"
)

func logFields(data map[string]interface{}, ev string, self, id uuid.UUID) logrus.Fields {
	d := make(logrus.Fields, len(data)+2)
	for k, v := range data {
		d[k] = v
	}

	if id != uuid.Nil {
		d[logger.DefaultTargetIDKey] = id
	}

	if self != uuid.Nil {
		d[logger.DefaultUserIDKey] = self
	}

	d[logger.DefaultEventKey] = ev

	return d
}
