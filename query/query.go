package query

import (
	"errors"
	"fmt"
	"github.com/lib/pq"
	"net/url"
	"strconv"
	"strings"
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
)

const (
	OrderAsc  = "asc"
	OrderDesc = "desc"
)

type Expr struct {
	Col   string
	Op    string
	Value string
}

type Query struct {
	SortBy string
	Order  string
	Last   string
	LastID string
	Limit  int
	Match  []Expr
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
}

func (e *Expr) Expr(val string) string {
	col := pq.QuoteIdentifier(e.Col)

	switch e.Op {
	case OpEq:
		return col + " = " + val
	case OpNe:
		return col + " <> " + val
	case OpLT:
		return col + " < " + val
	case OpGT:
		return col + " > " + val
	case OpLE:
		return col + " <= " + val
	case OpGE:
		return col + " >= " + val
	case OpRegex:
		return col + " ~ " + val
	case OpLike:
		return col + " LIKE " + val
	case OpPrefix:
		return col + " LIKE (" + val + " || '%')"
	case OpSuffix:
		return col + " LIKE ('%' || " + val + ")"
	case OpSubstr:
		return col + " LIKE ('%' || " + val + " || '%')"
	default:
		return ""
	}
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

			if _, ok := validOp[e.Op]; !ok {
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

	if q.Limit != 0 {
		ret.Set("limit", strconv.FormatInt(int64(q.Limit), 10))
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
	Table          string
	IDColumn       string
	ReturnColumn   string
	ValidateColumn func(string) bool
}

func (q *Query) SelectStmt(o *SelectOptions) (string, []interface{}, error) {
	if q.SortBy == "" {
		return "", nil, errors.New("Sorting column is not specified")
	}

	if o.ValidateColumn != nil {
		if err := q.ValidateColumnsFunc(o.ValidateColumn); err != nil {
			return "", nil, err
		}
	}

	sortCol := pq.QuoteIdentifier(q.SortBy)

	expr := "SELECT *"
	if o.ReturnColumn != "" {
		expr += ", " + sortCol + " AS " + pq.QuoteIdentifier(o.ReturnColumn)
	}
	expr += " FROM " + pq.QuoteIdentifier(o.Table)

	var i argIndex
	arg := make([]interface{}, 0, len(q.Match)+1)
	idCol := pq.QuoteIdentifier(o.IDColumn)

	var cond string

	if q.Last != "" {
		if q.LastID != "" {
			cond = "("
		}

		lastArg := i.Next()
		cond += sortCol + " > " + lastArg
		arg = append(arg, q.Last)

		if q.LastID != "" {
			cond += " OR " + sortCol + " = " + lastArg + " AND " + idCol + " > " + i.Next() + ")"
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

	expr += " ORDER BY " + sortCol + " " + so + ", " + idCol + " ASC"

	if q.Limit > 0 {
		expr += " LIMIT " + i.Next()
		arg = append(arg, q.Limit)
	}

	return expr, arg, nil
}

func errCol(col string) error {
	return fmt.Errorf("Invalid column name `%s'", col)
}

func (q *Query) ValidateColumnsFunc(f func(string) bool) error {
	if !f(q.SortBy) {
		return errCol(q.SortBy)
	}

	for _, e := range q.Match {
		if !f(e.Col) {
			return errCol(e.Col)
		}
	}

	return nil
}
