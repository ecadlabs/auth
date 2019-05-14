package jq

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExpr(t *testing.T) {
	tests := []struct {
		u   url.Values
		o   *Options
		sql string
	}{
		{
			u: url.Values{
				"q":      []string{`{"and":[{"eq":{"A":"v0"}},{"eq":{"B":"v1"}}]}`},
				"sortBy": []string{"C"},
				"last":   []string{"cval"},
				"lastId": []string{"idval"},
				"limit":  []string{"10"},
			},
			o: &Options{
				IDColumn: "id",
				FromExpr: "FROM table",
			},
			sql: "SELECT * FROM table WHERE ((A = $1) AND (B = $2)) AND ((C > $3) OR ((C = $4) AND (id > $5))) ORDER BY C ASC, id ASC LIMIT $6",
		},
		{
			u: url.Values{
				"q":      []string{`{"and":[{"eq":{"A":"v0"}},{"eq":{"B":"v1"}}]}`},
				"sortBy": []string{"C"},
				"last":   []string{"cval"},
			},
			o: &Options{
				IDColumn: "id",
				FromExpr: "FROM table",
			},
			sql: "SELECT * FROM table WHERE ((A = $1) AND (B = $2)) AND (C > $3) ORDER BY C ASC, id ASC",
		},
		{
			u: url.Values{
				"q": []string{`{"and":[{"eq":{"A":"v0"}},{"eq":{"B":"v1"}}]}`},
			},
			o: &Options{
				IDColumn: "id",
				FromExpr: "FROM table",
			},
			sql: "SELECT * FROM table WHERE (A = $1) AND (B = $2) ORDER BY id ASC",
		},
		{
			u: url.Values{},
			o: &Options{
				IDColumn: "id",
				FromExpr: "FROM table",
			},
			sql: "SELECT * FROM table ORDER BY id ASC",
		},
		{
			u: url.Values{
				"sortBy": []string{"C"},
				"last":   []string{"cval"},
				"lastId": []string{"idval"},
			},
			o: &Options{
				IDColumn: "id",
				FromExpr: "FROM table",
			},
			sql: "SELECT * FROM table WHERE (C > $1) OR ((C = $2) AND (id > $3)) ORDER BY C ASC, id ASC",
		},
	}

	for _, tst := range tests {
		query, err := FromValues(tst.u)
		assert.Nil(t, err)

		stmt, _, err := query.SelectStmt(tst.o)
		assert.Nil(t, err)
		assert.Equal(t, tst.sql, stmt)
	}
}
