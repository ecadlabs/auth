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
	DefaultSourceIDKey = "source_id"
	TargetIDType       = "target_id_type"
	SourceIDType       = "source_id_type"
	DefaultTargetIDKey = "id"
	DefaultEventKey    = "event"
	DefaultAddrKey     = "addr"
)

var hookLevels = []logrus.Level{logrus.InfoLevel}

type Hook struct {
	DB        *sql.DB
	Table     string
	UserIDKey string
	// TenantIDKey       string
	TargetIDKey string
	// TargetTenantIDKey string
	EventKey string
	AddrKey  string
}

func (h *Hook) table() string {
	if h.Table != "" {
		return h.Table
	}

	return DefaultTable
}

func (h *Hook) sourceIDKey() string {
	if h.UserIDKey != "" {
		return h.UserIDKey
	}

	return DefaultSourceIDKey
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

	uid := entry.Data[h.sourceIDKey()]
	sourceIdType := entry.Data[SourceIDType]
	targetIdType := entry.Data[TargetIDType]
	tid := entry.Data[h.targetIDKey()]
	ev := entry.Data[h.eventKey()]
	addr := entry.Data[h.addrKey()]

	_, err = h.DB.Exec("INSERT INTO "+pq.QuoteIdentifier(h.table())+" (id, ts, event, source_id, target_id, source_type, target_type, addr, msg, data) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)", uuid.NewV4(), entry.Time, ev, uid, tid, sourceIdType, targetIdType, addr, entry.Message, buf)
	if err != nil {
		return err
	}

	return nil
}
