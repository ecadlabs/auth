package query

import (
	"net/url"
	"sort"
	"testing"
)

func TestQuery(t *testing.T) {
	tests := []struct {
		q    string
		sql  string
		args []interface{}
	}{
		{
			q:    "sortBy=sort_col",
			sql:  "SELECT * FROM \"table\" ORDER BY \"sort_col\" ASC, \"id\" ASC",
			args: []interface{}{},
		},
		{
			q:    "sortBy=sort_col&limit=10",
			sql:  "SELECT * FROM \"table\" ORDER BY \"sort_col\" ASC, \"id\" ASC LIMIT $1",
			args: []interface{}{int(10)},
		},
		{
			q:    "sortBy=sort_col&limit=10&last=start_val",
			sql:  "SELECT * FROM \"table\" WHERE \"sort_col\" > $1 ORDER BY \"sort_col\" ASC, \"id\" ASC LIMIT $2",
			args: []interface{}{"start_val", int(10)},
		},
		{
			q:    "sortBy=sort_col&limit=10&last=start_val&col1[eq]=val1&col2[p]=val2",
			sql:  "SELECT * FROM \"table\" WHERE \"sort_col\" > $1 AND \"col1\" = $2 AND \"col2\" LIKE ($3 || '%') ORDER BY \"sort_col\" ASC, \"id\" ASC LIMIT $4",
			args: []interface{}{"start_val", "val1", "val2", int(10)},
		},
		{
			q:    "sortBy=sort_col&limit=10&last=start_val&lastId=id_val&col1[eq]=val1&col2[p]=val2",
			sql:  "SELECT * FROM \"table\" WHERE (\"sort_col\" > $1 OR \"sort_col\" = $1 AND \"id\" > $2) AND \"col1\" = $3 AND \"col2\" LIKE ($4 || '%') ORDER BY \"sort_col\" ASC, \"id\" ASC LIMIT $5",
			args: []interface{}{"start_val", "id_val", "val1", "val2", int(10)},
		},
		{
			q:    "sortBy=sort_col&col1[eq]=val1&col2[p]=val2",
			sql:  "SELECT * FROM \"table\" WHERE \"col1\" = $1 AND \"col2\" LIKE ($2 || '%') ORDER BY \"sort_col\" ASC, \"id\" ASC",
			args: []interface{}{"val1", "val2"},
		},
	}

	for _, tst := range tests {
		u, err := url.ParseQuery(tst.q)
		if err != nil {
			t.Error(err)
			continue
		}

		q, err := FromValues(u)
		if err != nil {
			t.Error(err)
			continue
		}

		sort.Slice(q.Match, func(i, j int) bool { return q.Match[i].Col < q.Match[j].Col })

		selOpt := SelectOptions{
			FromExpr: "\"table\"",
			IDColumn: "id",
		}

		sql, args, err := q.SelectStmt(&selOpt)
		if err != nil {
			t.Error(err)
			continue
		}

		/*
			fmt.Println(sql)
			fmt.Printf("%#v\n", args)
		*/

		if sql != tst.sql {
			t.Errorf("'%s' != '%s'", sql, tst.sql)
			continue
		}

		if len(args) != len(tst.args) {
			t.Errorf("%d != %d", len(args), len(tst.args))
			continue
		}

		for i, v := range args {
			if v != tst.args[i] {
				t.Errorf("%v != %v", v, tst.args[i])
			}
		}
	}
}
