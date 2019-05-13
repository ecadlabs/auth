package query

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

const (
	OpEq             = "eq"
	OpNe             = "ne"
	OpLT             = "lt"
	OpGT             = "gt"
	OpLE             = "le"
	OpGE             = "ge"
	OpRegex          = "re"
	OpRegexSensitive = "res"
	OpLike           = "l"
	OpPrefix         = "p"
	OpSuffix         = "s"
	OpSubstr         = "sub"
	OpHas            = "has"
)

const (
	OrderAsc  = "asc"
	OrderDesc = "desc"
)

const (
	ColSort = 1 << iota
	ColNull
	ColNumeric
)

type Expr struct {
	Col   string
	Op    string
	Value string
}

type ColumnFunc func(string) (column string, flags int)

type PositionalArgFunc func(int) string

type ColumnOptions struct {
	Name  string
	Flags int
}

type Columns map[string]ColumnOptions

func (c Columns) Func(name string) (column string, flags int) {
	if o, ok := c[name]; ok {
		return o.Name, o.Flags
	}
	return "", 0
}

type Query struct {
	SortBy     string
	Order      string
	Last       *string
	LastID     *string
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

func ColumnExpr(col string, fn ColumnFunc) (string, error) {
	if fn == nil {
		return col, nil
	}

	res, flags := fn(col)
	if res == "" {
		return "", fmt.Errorf("Unknown column `%s'", col)
	}

	// Emit COALESCE expression
	if flags&ColNull != 0 {
		var def string

		if flags&ColNumeric != 0 {
			def = "0"
		} else {
			def = "''"
		}

		return "COALESCE(" + res + ", " + def + ")", nil
	}

	return res, nil
}

func (e *Expr) expr(val string, fn ColumnFunc) (string, error) {
	var neg bool
	op := e.Op
	if len(op) != 0 && op[0] == '!' {
		op = op[1:]
		neg = true
	}

	col, err := ColumnExpr(e.Col, fn)
	if err != nil {
		return "", err
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
		expr = col + " ~* " + val
	case OpRegexSensitive:
		expr = col + " ~ " + val
	case OpLike:
		expr = col + " LIKE " + val
	case OpPrefix:
		expr = col + " LIKE CONCAT(" + val + ", '%')"
	case OpSuffix:
		expr = col + " LIKE CONCAT('%', " + val + ")"
	case OpSubstr:
		expr = col + " LIKE CONCAT('%', " + val + ", '%')"
	case OpHas:
		expr = "(" + col + " IS NOT NULL AND " + val + " = ANY(" + col + "))"
	default:
		expr = col + " = " + val
	}

	if neg {
		return "NOT " + expr, nil
	}

	return expr, nil
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
				res.Last = &v
			case "lastId":
				res.LastID = &v
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

	for _, e := range q.Match {
		ret.Set(e.Col+"["+e.Op+"]", e.Value)
	}

	return ret
}

type SelectOptions struct {
	SelectExpr        string
	FromExpr          string
	IDColumn          string
	ColumnFunc        ColumnFunc
	PositionalArgFunc PositionalArgFunc
}

func PostgresPositional(i int) string { return "$" + strconv.FormatInt(int64(i), 10) }
func MySQLPositional(i int) string    { return "?" }

func (s *SelectOptions) pos(i int) string {
	if s.PositionalArgFunc != nil {
		return s.PositionalArgFunc(i)
	}

	return PostgresPositional(i)
}

type argIndex struct {
	i int
	o *SelectOptions
}

func (a *argIndex) next() string {
	a.i++
	return a.o.pos(a.i)
}

func (q *Query) CountStmt(o *SelectOptions) (string, []interface{}, error) {
	stmt := "SELECT COUNT(*) FROM " + o.FromExpr
	arg := make([]interface{}, len(q.Match))

	idx := argIndex{o: o}
	var cond string

	for i, m := range q.Match {
		if cond != "" {
			cond += " AND "
		}

		ex, err := m.expr(idx.next(), o.ColumnFunc)
		if err != nil {
			return "", nil, err
		}

		cond += ex
		arg[i] = m.Value
	}

	if len(q.Match) != 0 {
		stmt += " WHERE " + cond
	}

	return stmt, arg, nil
}

func (q *Query) SelectStmt(o *SelectOptions) (string, []interface{}, error) {
	var (
		sortBy  string // Input property name
		sortCol string // Resulting column name
	)

	if q.SortBy == "" {
		if o.IDColumn != "" {
			sortBy = o.IDColumn
		} else {
			return "", nil, errors.New("Sorting column is not specified")
		}
	} else {
		sortBy = q.SortBy
	}

	if o.ColumnFunc != nil {
		var flags int
		if sortCol, flags = o.ColumnFunc(sortBy); sortCol == "" || flags&ColSort == 0 {
			return "", nil, fmt.Errorf("Can't sort by column `%s'", sortBy)
		}
	} else {
		sortCol = sortBy
	}

	se := o.SelectExpr
	if se == "" {
		se = "*"
	}

	expr := "SELECT " + se + " FROM " + o.FromExpr

	i := argIndex{o: o}
	arg := make([]interface{}, 0, len(q.Match)+1)

	var cmp string
	if q.Order == OrderDesc {
		cmp = "<"
	} else {
		cmp = ">"
	}

	var cond string

	if q.Last != nil {
		extendedExpr := q.LastID != nil && sortCol != o.IDColumn

		if extendedExpr {
			cond = "("
		}

		cond += sortCol + " " + cmp + " " + i.next()
		arg = append(arg, *q.Last)

		if extendedExpr {
			cond += " OR " + sortCol + " = " + i.next() + " AND " + o.IDColumn + " " + cmp + " " + i.next() + ")"
			arg = append(arg, *q.Last, *q.LastID)
		}
	}

	for _, m := range q.Match {
		if cond != "" {
			cond += " AND "
		}

		ex, err := m.expr(i.next(), o.ColumnFunc)
		if err != nil {
			return "", nil, err
		}

		cond += ex
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

	expr += " ORDER BY " + sortCol + " " + so
	if sortCol != o.IDColumn {
		expr += ", " + o.IDColumn + " " + so
	}

	if q.Limit > 0 {
		expr += " LIMIT " + i.next()
		arg = append(arg, q.Limit)
	}

	return expr, arg, nil
}

func errCol(col string) error {
	return fmt.Errorf("Invalid column name `%s'", col)
}
