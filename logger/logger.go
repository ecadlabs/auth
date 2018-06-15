package logger

import (
	"database/sql"
	"github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
)

const (
	EvCreate     = "create"
	EvUpdate     = "update"
	EvAddRole    = "add_role"
	EvRemoveRole = "remove_role"
	EvDelete     = "delete"
)

type Logger struct {
	*logrus.Logger
	hook *Hook
}

func NewWithHook(h *Hook) *Logger {
	l := logrus.New()
	l.AddHook(h)

	return &Logger{
		Logger: l,
		hook:   h,
	}
}

func New(db *sql.DB) *Logger {
	return NewWithHook(&Hook{DB: db})
}

func (l *Logger) fields(data map[string]interface{}, ev string, self, id uuid.UUID) logrus.Fields {
	d := make(logrus.Fields, len(data)+2)
	for k, v := range data {
		d[k] = v
	}

	if id != uuid.Nil {
		d[l.hook.targetIDKey()] = id
	}

	if self != uuid.Nil {
		d[l.hook.userIDKey()] = self
	}

	d[l.hook.eventKey()] = ev

	return d
}

func (l *Logger) Created(self, id uuid.UUID, data map[string]interface{}) {
	l.WithFields(l.fields(data, EvCreate, self, id)).Printf("User %v created account %v", self, id)
}

func (l *Logger) Updated(self, id uuid.UUID, data map[string]interface{}) {
	l.WithFields(l.fields(data, EvUpdate, self, id)).Printf("User %v updated account %v", self, id)
}

func (l *Logger) Deleted(self, id uuid.UUID) {
	l.WithFields(l.fields(nil, EvDelete, self, id)).Printf("User %v deleted account %v", self, id)
}

func (l *Logger) RoleAdded(self, id uuid.UUID, role string) {
	l.WithFields(l.fields(map[string]interface{}{"role": role}, EvAddRole, self, id)).Printf("User %v added role `%s' to account %v", self, role, id)
}

func (l *Logger) RoleRemoved(self, id uuid.UUID, role string) {
	l.WithFields(l.fields(map[string]interface{}{"role": role}, EvRemoveRole, self, id)).Printf("User %v removed role `%s' from account %v", self, role, id)
}
