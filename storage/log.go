package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ecadlabs/auth/errors"
	"github.com/ecadlabs/auth/jq"
	uuid "github.com/satori/go.uuid"
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
	SortedBy   string         `db:"_sorted_by"` // Output only
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

var logQueryColumns = jq.Columns{
	"id":        {ColumnName: "id", Sort: true},
	"ts":        {ColumnName: "ts", Sort: true},
	"event":     {ColumnName: "event", Sort: true},
	"source_id": {ColumnName: "source_id", Sort: true},
	"target_id": {ColumnName: "target_id", Sort: true},
	"addr":      {ColumnName: "addr", Sort: true},
}

// GetLogs retrive logs from the database as a paged results
func (s *Storage) GetLogs(ctx context.Context, query *jq.Query) (entries []*LogEntry, count int, next *jq.Query, err error) {
	q := *query

	if q.SortBy == "" {
		q.SortBy = LogDefaultSortColumn
	}

	sortExpr, err := jq.ColumnExpr(q.SortBy, logQueryColumns)
	if err != nil {
		err = errors.Wrap(err, errors.CodeQuerySyntax)
		return
	}

	selOpt := jq.Options{
		SelectExpr:   fmt.Sprintf("SELECT *, %s AS _sorted_by", sortExpr),
		FromExpr:     "FROM log",
		IDColumn:     "id",
		Columns:      logQueryColumns,
		DriverParams: jq.PostgresDriverParams,
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
		lastID := lastItem.ID.String()
		ret := *query
		ret.LastID = &lastID
		ret.Last = &lastItem.SortedBy
		ret.TotalCount = false

		next = &ret
	}

	return
}
