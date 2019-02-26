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
	ID         uuid.UUID      `db:"id"`
	Timestamp  time.Time      `db:"ts"`
	Event      string         `db:"event"`
	SourceID   uuid.UUID      `db:"source_id"`
	TargerID   uuid.UUID      `db:"target_id"`
	SourceType string         `db:"source_type"`
	TargetType string         `db:"target_type"`
	Data       []byte         `db:"data"`
	Address    string         `db:"addr"`
	Message    sql.NullString `db:"msg"`
	SortedBy   string         `db:"sorted_by"` // Output only
}

func (l *logEntryModel) toLogEntry() *LogEntry {
	ret := &LogEntry{
		ID:         l.ID,
		Timestamp:  l.Timestamp,
		Event:      l.Event,
		SourceID:   l.SourceID,
		TargerID:   l.TargerID,
		SourceType: l.SourceType,
		TargetType: l.TargetType,
		Address:    l.Address,
		Message:    l.Message.String,
	}

	if len(l.Data) != 0 {
		if err := json.Unmarshal(l.Data, &ret.Data); err != nil {
			log.Error(err)
		}
	}

	return ret
}

var logQueryColumns = query.Columns{
	"ts":        {Name: "ts", Flags: query.ColSort},
	"event":     {Name: "event", Flags: query.ColSort},
	"source_id": {Name: "source_id", Flags: query.ColSort},
	"target_id": {Name: "target_id", Flags: query.ColSort},
	"addr":      {Name: "addr", Flags: query.ColSort},
}

// GetLogs retrive logs from the database as a paged results
func (s *Storage) GetLogs(ctx context.Context, q *query.Query) (entries []*LogEntry, count int, next *query.Query, err error) {
	if q.SortBy == "" {
		q.SortBy = LogDefaultSortColumn
	}

	selOpt := query.SelectOptions{
		SelectExpr: "*, " + pq.QuoteIdentifier(q.SortBy) + " AS sorted_by",
		FromExpr:   "log",
		IDColumn:   "id",
		ColumnFunc: logQueryColumns.Func,
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
		if stmt, args, err = q.CountStmt(&selOpt); err != nil {
			return
		}

		if err = s.DB.Get(&count, stmt, args...); err != nil {
			return
		}
	}

	entries = logSlice

	if lastItem != nil {
		// Update query
		lastid := lastItem.ID.String()
		ret := *q
		ret.LastID = &lastid
		ret.Last = &lastItem.SortedBy
		ret.TotalCount = false

		next = &ret
	}

	return
}
