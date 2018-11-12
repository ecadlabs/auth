package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/ecadlabs/auth/errors"
	"github.com/ecadlabs/auth/query"
	"github.com/lib/pq"
	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

type logEntryModel struct {
	ID        uuid.UUID      `db:"id"`
	Timestamp time.Time      `db:"ts"`
	Event     string         `db:"event"`
	UserID    uuid.UUID      `db:"user_id"`
	TargerID  uuid.UUID      `db:"target_id"`
	Data      []byte         `db:"data"`
	Address   string         `db:"addr"`
	Message   sql.NullString `db:"msg"`
	SortedBy  string         `db:"sorted_by"` // Output only
}

func (l *logEntryModel) toLogEntry() *LogEntry {
	ret := &LogEntry{
		ID:        l.ID,
		Timestamp: l.Timestamp,
		Event:     l.Event,
		UserID:    l.UserID,
		TargerID:  l.TargerID,
		Address:   l.Address,
		Message:   l.Message.String,
	}

	if len(l.Data) != 0 {
		if err := json.Unmarshal(l.Data, &ret.Data); err != nil {
			log.Error(err)
		}
	}

	return ret
}

var logQueryColumns = map[string]int{
	"ts":        query.ColQuery | query.ColSort,
	"event":     query.ColQuery | query.ColSort,
	"user_id":   query.ColQuery | query.ColSort,
	"target_id": query.ColQuery | query.ColSort,
	"addr":      query.ColQuery | query.ColSort,
}

func (s *Storage) GetLogs(ctx context.Context, q *query.Query) (entries []*LogEntry, count int, next *query.Query, err error) {
	if q.SortBy == "" {
		q.SortBy = LogDefaultSortColumn
	}

	selOpt := query.SelectOptions{
		SelectExpr: "*, " + pq.QuoteIdentifier(q.SortBy) + " AS sorted_by",
		FromExpr:   "log",
		IDColumn:   "id",
		ColumnFlagsFunc: func(col string) int {
			if flags, ok := logQueryColumns[col]; ok {
				return flags
			}
			return 0
		},
	}

	stmt, args, err := q.SelectStmt(&selOpt)
	if err != nil {
		err = errors.Wrap(err, errors.CodeQuerySyntax)
	}

	rows, err := s.DB.QueryxContext(ctx, stmt, args...)
	if err != nil {
		return
	}
	defer rows.Close()

	logSlice := []*LogEntry{}
	var lastItem *logEntryModel

	for rows.Next() {
		var le logEntryModel
		if err = rows.StructScan(&le); err != nil {
			return
		}

		lastItem = &le
		logSlice = append(logSlice, le.toLogEntry())
	}

	if err = rows.Err(); err != nil {
		return
	}

	// Count
	if q.TotalCount {
		stmt, args := q.CountStmt(&selOpt)
		if err = s.DB.Get(&count, stmt, args...); err != nil {
			return
		}
	}

	entries = logSlice

	if lastItem != nil {
		// Update query
		ret := *q
		ret.LastID = lastItem.ID.String()
		ret.Last = lastItem.SortedBy
		ret.TotalCount = false

		next = &ret
	}

	return
}
