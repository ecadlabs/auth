package jq

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

type Options struct {
	SelectExpr   string
	FromExpr     string
	IDColumn     string
	Columns      Columns
	DriverParams *DriverParams
}

func (o *Options) driverParams() *DriverParams {
	if o.DriverParams != nil {
		return o.DriverParams
	}
	return DefaultDriverParams
}

type Query struct {
	SortBy     string
	Order      string
	Last       *string
	LastID     *string
	Limit      int
	TotalCount bool
	Expr       *Expr
	RawExpr    string
}

const (
	OrderAsc  = "asc"
	OrderDesc = "desc"
)

var validOrder = map[string]struct{}{
	OrderAsc:  struct{}{},
	OrderDesc: struct{}{},
}

func FromValues(q url.Values) (*Query, error) {
	var res Query

	// Query itself
	if str := q.Get("q"); str != "" {
		var expr Expr
		if err := json.Unmarshal([]byte(str), &expr); err != nil {
			return nil, err
		}
		res.Expr = &expr
		res.RawExpr = str
	}

	res.SortBy = q.Get("sortBy")

	if str := q.Get("last"); str != "" {
		res.Last = &str
	}

	if str := q.Get("lastId"); str != "" {
		res.LastID = &str
	}

	if str := q.Get("order"); str != "" {
		if _, ok := validOrder[str]; !ok {
			return nil, fmt.Errorf("Incorrect sorting order: `%s'", str)
		}
		res.Order = str
	}

	if str := q.Get("limit"); str != "" {
		i, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return nil, err
		}
		res.Limit = int(i)
	}

	if str := q.Get("count"); str != "" {
		b, err := strconv.ParseBool(str)
		if err != nil {
			return nil, err
		}
		res.TotalCount = b
	}

	return &res, nil
}

func (q *Query) Values() url.Values {
	ret := make(url.Values)

	if q.RawExpr != "" {
		ret.Set("q", q.RawExpr)
	}

	if q.SortBy != "" {
		ret.Set("sortBy", q.SortBy)
	}

	if q.Order != "" {
		ret.Set("order", q.Order)
	}

	if q.Last != nil {
		ret.Set("last", *q.Last)
	}

	if q.LastID != nil {
		ret.Set("lastId", *q.LastID)
	}

	if q.Limit > 0 {
		ret.Set("limit", strconv.FormatInt(int64(q.Limit), 10))
	}

	if q.TotalCount {
		ret.Set("count", "1")
	}

	return ret
}

func (q *Query) CountStmt(o *Options) (stmt string, arg []interface{}, err error) {
	var s strings.Builder
	s.WriteString("SELECT COUNT(*) " + o.FromExpr)

	var idx int
	if q.Expr != nil {
		var expr string
		expr, arg, err = q.Expr.SQL(&idx, o.driverParams(), o.Columns)
		if err != nil {
			return
		}
		s.WriteString(" WHERE " + expr)
	}
	stmt = s.String()
	return
}

func orderNode(order, key string, val interface{}) Node {
	if order == OrderDesc {
		return &LTExpr{key, val}
	} else {
		return &GTExpr{key, val}
	}
}

func (q *Query) SelectStmt(o *Options) (string, []interface{}, error) {
	sortBy := q.SortBy

	if q.SortBy == "" {
		if o.IDColumn != "" {
			sortBy = o.IDColumn
		} else {
			return "", nil, errors.New("Sorting column is not specified")
		}
	}

	sortSQLColumn := sortBy
	idSQLColumn := o.IDColumn

	if o.Columns != nil {
		c, ok := o.Columns[sortBy]
		if !ok || c == nil || !c.Sort {
			return "", nil, fmt.Errorf("Can't sort by column `%s'", sortBy)
		}
		sortSQLColumn = c.ColumnExpr

		idC, idOk := o.Columns[o.IDColumn]
		if !idOk || idC == nil {
			return "", nil, fmt.Errorf("Unknown column `%s'", o.IDColumn)
		}
		idSQLColumn = idC.ColumnExpr
	}

	expr := q.Expr
	if q.Last != nil {
		// Append pagination expression
		var node Node
		if q.LastID != nil && sortBy != o.IDColumn {
			node = &ORExpr{
				&Expr{Node: orderNode(q.Order, sortBy, *q.Last)},
				&Expr{Node: &ANDExpr{
					&Expr{Node: &EQExpr{sortBy, *q.Last}},
					&Expr{Node: orderNode(q.Order, o.IDColumn, *q.LastID)},
				}},
			}
		} else {
			node = orderNode(q.Order, sortBy, *q.Last)
		}

		if expr == nil {
			expr = &Expr{Node: node}
		} else {
			expr = &Expr{Node: &ANDExpr{
				expr,
				&Expr{Node: node},
			}}
		}
	}

	se := o.SelectExpr
	if se == "" {
		se = "SELECT *"
	}

	var (
		index int
		args  []interface{}
	)

	var stmt strings.Builder
	stmt.WriteString(se + " " + o.FromExpr)

	if expr != nil {
		e, a, err := expr.SQL(&index, o.driverParams(), o.Columns)
		if err != nil {
			return "", nil, err
		}
		args = a
		stmt.WriteString(" WHERE " + e)
	}

	var so string
	if q.Order == OrderDesc {
		so = "DESC"
	} else {
		so = "ASC"
	}

	stmt.WriteString(" ORDER BY " + sortSQLColumn + " " + so)
	if sortBy != o.IDColumn {
		stmt.WriteString(", " + idSQLColumn + " " + so)
	}

	if q.Limit > 0 {
		index++
		stmt.WriteString(" LIMIT " + o.driverParams().pos(index))
		args = append(args, q.Limit)
	}

	return stmt.String(), args, nil
}
