package query

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/lib/pq"
)

const (
	OpEq     = "eq"
	OpNe     = "ne"
	OpLT     = "lt"
	OpGT     = "gt"
	OpLE     = "le"
	OpGE     = "ge"
	OpRegex  = "re"
	OpLike   = "l"
	OpPrefix = "p"
	OpSuffix = "s"
	OpSubstr = "sub"
	OpHas    = "has"
)

const (
	OrderAsc  = "asc"
	OrderDesc = "desc"
)

const (
	ColQuery = 1 << iota
	ColSort
)

type Expr struct {
	Col   string
	Op    string
	Value string
}

type ColFlagsFunc func(string) int

type Query struct {
	SortBy     string
	Order      string
	Last       string
	LastID     string
	Limit      int
	TotalCount bool
	Match      []Expr
}

var validOrder = map[string]struct{}{
	OrderAsc:  struct{}{},
	OrderDesc: struct{}{},
}

var validOp = map[string]struct{}{
	OpEq:     struct{}{},
	OpNe:     struct{}{},
	OpLT:     struct{}{},
	OpGT:     struct{}{},
	OpLE:     struct{}{},
	OpGE:     struct{}{},
	OpRegex:  struct{}{},
	OpLike:   struct{}{},
	OpPrefix: struct{}{},
	OpSuffix: struct{}{},
	OpSubstr: struct{}{},
	OpHas:    struct{}{},
}

func (e *Expr) Expr(val string) string {
	col := pq.QuoteIdentifier(e.Col)

	var neg bool
	op := e.Op
	if len(op) != 0 && op[0] == '!' {
		op = op[1:]
		neg = true
	}

	var expr string

	switch op {
	case OpNe:
		expr = col + " <> " + val
	case OpLT:
		expr = col + " < " + val
	case OpGT:
		expr = col + " > " + val
	case OpLE:
		expr = col + " <= " + val
	case OpGE:
		expr = col + " >= " + val
	case OpRegex:
		expr = col + " ~ " + val
	case OpLike:
		expr = col + " LIKE " + val
	case OpPrefix:
		expr = col + " LIKE (" + val + " || '%')"
	case OpSuffix:
		expr = col + " LIKE ('%' || " + val + ")"
	case OpSubstr:
		expr = col + " LIKE ('%' || " + val + " || '%')"
	case OpHas:
		expr = "(" + col + " IS NOT NULL AND " + val + " = ANY(" + col + "))"
	default:
		expr = col + " = " + val
	}

	if neg {
		return "NOT " + expr
	}

	return expr
}

func FromValues(q url.Values) (*Query, error) {
	var res Query

	for k, val := range q {
		if len(val) == 0 {
			continue
		}

		v := val[0]

		start := strings.IndexByte(k, '[')
		end := strings.IndexByte(k, ']')

		if start > 0 && end >= start {
			e := Expr{
				Col:   k[:start],
				Op:    k[start+1 : end],
				Value: v,
			}

			op := e.Op
			if len(op) != 0 && op[0] == '!' {
				op = op[1:]
			}

			if _, ok := validOp[op]; !ok {
				return nil, fmt.Errorf("Incorrect operator: `%s'", e.Op)
			}

			res.Match = append(res.Match, e)
		} else if start < 0 && end < 0 {
			switch k {
			case "sortBy":
				res.SortBy = v
			case "last":
				res.Last = v
			case "lastId":
				res.LastID = v
			case "order":
				if _, ok := validOrder[v]; !ok {
					return nil, fmt.Errorf("Incorrect order: `%s'", v)
				}
				res.Order = v
			case "limit":
				i, err := strconv.ParseInt(v, 10, 64)
				if err != nil {
					return nil, err
				}
				res.Limit = int(i)
			case "count":
				b, err := strconv.ParseBool(v)
				if err != nil {
					return nil, err
				}
				res.TotalCount = b
			}
		} else {
			return nil, fmt.Errorf("Incorrect query: `%s'", k)
		}
	}

	return &res, nil
}

func (q *Query) Values() url.Values {
	ret := make(url.Values)

	if q.SortBy != "" {
		ret.Set("sortBy", q.SortBy)
	}

	if q.Order != "" {
		ret.Set("order", q.Order)
	}

	if q.Last != "" {
		ret.Set("last", q.Last)
	}

	if q.LastID != "" {
		ret.Set("lastId", q.LastID)
	}

	if q.Limit > 0 {
		ret.Set("limit", strconv.FormatInt(int64(q.Limit), 10))
	}

	if q.TotalCount {
		ret.Set("count", "1")
	}

	for _, e := range q.Match {
		ret.Set(e.Col+"["+e.Op+"]", e.Value)
	}

	return ret
}

type argIndex int

func (a *argIndex) Next() string {
	(*a)++
	return fmt.Sprintf("$%d", *a)
}

type SelectOptions struct {
	SelectExpr      string
	FromExpr        string
	IDColumn        string
	ColumnFlagsFunc ColFlagsFunc
}

func (q *Query) CountStmt(o *SelectOptions) (string, []interface{}) {
	stmt := "SELECT COUNT(*) FROM " + o.FromExpr
	arg := make([]interface{}, len(q.Match))

	var (
		idx  argIndex
		cond string
	)

	for i, m := range q.Match {
		if cond != "" {
			cond += " AND "
		}

		cond += m.Expr(idx.Next())
		arg[i] = m.Value
	}

	if len(q.Match) != 0 {
		stmt += " WHERE " + cond
	}

	return stmt, arg
}

func (q *Query) SelectStmt(o *SelectOptions) (string, []interface{}, error) {
	if q.SortBy == "" {
		return "", nil, errors.New("Sorting column is not specified")
	}

	if o.ColumnFlagsFunc != nil {
		if err := q.validateColumnsFunc(o.ColumnFlagsFunc); err != nil {
			return "", nil, err
		}
	}

	sortCol := pq.QuoteIdentifier(q.SortBy)

	se := o.SelectExpr
	if se == "" {
		se = "*"
	}

	expr := "SELECT " + se + " FROM " + o.FromExpr

	var i argIndex
	arg := make([]interface{}, 0, len(q.Match)+1)
	idCol := pq.QuoteIdentifier(o.IDColumn)

	var cmp string
	if q.Order == OrderDesc {
		cmp = "<"
	} else {
		cmp = ">"
	}

	var cond string

	if q.Last != "" {
		if q.LastID != "" {
			cond = "("
		}

		lastArg := i.Next()
		cond += sortCol + " " + cmp + " " + lastArg
		arg = append(arg, q.Last)

		if q.LastID != "" {
			cond += " OR " + sortCol + " = " + lastArg + " AND " + idCol + " " + cmp + " " + i.Next() + ")"
			arg = append(arg, q.LastID)
		}
	}

	for _, m := range q.Match {
		if cond != "" {
			cond += " AND "
		}

		cond += m.Expr(i.Next())
		arg = append(arg, m.Value)
	}

	if cond != "" {
		expr += " WHERE " + cond
	}

	var so string
	if q.Order == OrderDesc {
		so = "DESC"
	} else {
		so = "ASC"
	}

	expr += " ORDER BY " + sortCol + " " + so + ", " + idCol + " " + so

	if q.Limit > 0 {
		expr += " LIMIT " + i.Next()
		arg = append(arg, q.Limit)
	}

	return expr, arg, nil
}

func errCol(col string) error {
	return fmt.Errorf("Invalid column name `%s'", col)
}

func (q *Query) validateColumnsFunc(f ColFlagsFunc) error {
	if f(q.SortBy)&ColSort == 0 {
		return errCol(q.SortBy)
	}

	for _, e := range q.Match {
		if f(e.Col)&ColQuery == 0 {
			return errCol(e.Col)
		}
	}

	return nil
}
