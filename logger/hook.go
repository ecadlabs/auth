package logger

import (
	"database/sql"
	"encoding/json"

	"github.com/lib/pq"
	"github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
)

const (
	DefaultTable       = "log"
	DefaultUserIDKey   = "user_id"
	DefaultTargetIDKey = "id"
	DefaultEventKey    = "event"
	DefaultAddrKey     = "addr"
)

var hookLevels = []logrus.Level{logrus.InfoLevel}

type Hook struct {
	DB          *sql.DB
	Table       string
	UserIDKey   string
	TargetIDKey string
	EventKey    string
	AddrKey     string
}

func (h *Hook) table() string {
	if h.Table != "" {
		return h.Table
	}

	return DefaultTable
}

func (h *Hook) userIDKey() string {
	if h.UserIDKey != "" {
		return h.UserIDKey
	}

	return DefaultUserIDKey
}

func (h *Hook) targetIDKey() string {
	if h.TargetIDKey != "" {
		return h.TargetIDKey
	}

	return DefaultTargetIDKey
}

func (h *Hook) eventKey() string {
	if h.EventKey != "" {
		return h.EventKey
	}

	return DefaultEventKey
}

func (h *Hook) addrKey() string {
	if h.AddrKey != "" {
		return h.AddrKey
	}

	return DefaultAddrKey
}

func (h *Hook) Levels() []logrus.Level {
	return hookLevels
}

func (h *Hook) Fire(entry *logrus.Entry) error {
	data := make(logrus.Fields, len(entry.Data))

	for k, v := range entry.Data {
		switch v := v.(type) {
		case error:
			data[k] = v.Error()
		default:
			data[k] = v
		}
	}

	buf, err := json.Marshal(data)
	if err != nil {
		return err
	}

	uid := entry.Data[h.userIDKey()]
	tid := entry.Data[h.targetIDKey()]
	ev := entry.Data[h.eventKey()]
	addr := entry.Data[h.addrKey()]

	_, err = h.DB.Exec("INSERT INTO "+pq.QuoteIdentifier(h.table())+" (id, ts, event, user_id, target_id, addr, msg, data) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", uuid.NewV4(), entry.Time, ev, uid, tid, addr, entry.Message, buf)
	if err != nil {
		return err
	}

	return nil
}
