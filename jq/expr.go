package jq

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type DriverParams struct {
	Pfn      PositionalArgFunc
	TextType string
}

type PositionalArgFunc func(int) string

const (
	PostgresTextType = "TEXT"
	MySQLTextType    = "CHAR"
)

var PostgresDriverParams = &DriverParams{
	Pfn:      PostgresPositional,
	TextType: PostgresTextType,
}

var MySQLDriverParams = &DriverParams{
	Pfn:      MySQLPositional,
	TextType: MySQLTextType,
}

func PostgresPositional(i int) string { return "$" + strconv.FormatInt(int64(i), 10) }
func MySQLPositional(i int) string    { return "?" }

var (
	DefaultDriverParams = PostgresDriverParams
	DefaultPositional   = PostgresPositional
	DefaultTextType     = PostgresTextType
)

type Node interface {
	SQL(index *int, p *DriverParams, columns Columns) (sql string, args []interface{}, err error)
}

type KeyVal struct {
	Key   string
	Value interface{}
}

type Column struct {
	ColumnName   string
	CoalesceExpr string
	Sort         bool
}

func (c *Column) Expr() string {
	if c.CoalesceExpr != "" {
		return c.CoalesceExpr
	}
	return c.ColumnName
}

type Columns map[string]*Column

func (k *KeyVal) UnmarshalJSON(data []byte) error {
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}

	if len(m) != 1 {
		return fmt.Errorf("Expression node must contain exactly 1 property, got %d instead", len(m))
	}

	for key, val := range m {
		k.Key = key
		k.Value = val
	}

	return nil
}

type Expr struct {
	Node
}

type List []*Expr

func (e *Expr) UnmarshalJSON(data []byte) error {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}

	if len(m) != 1 {
		return fmt.Errorf("Expression must contain exactly 1 property, got %d instead", len(m))
	}

	for key, val := range m {
		switch key {
		case "eq", "ne", "lt", "gt", "le", "ge", "re", "l", "p", "s", "sub", "has", "gg", "ll", "gge", "lle", "cnt":
			var child KeyVal
			if err := json.Unmarshal(val, &child); err != nil {
				return err
			}

			switch key {
			case "eq":
				e.Node = (*EQExpr)(&child)
			case "ne":
				e.Node = (*NEExpr)(&child)
			case "lt":
				e.Node = (*LTExpr)(&child)
			case "gt":
				e.Node = (*GTExpr)(&child)
			case "le":
				e.Node = (*LEExpr)(&child)
			case "ge":
				e.Node = (*GEExpr)(&child)
			case "re":
				e.Node = (*ReExpr)(&child)
			case "l":
				e.Node = (*LExpr)(&child)
			case "p":
				e.Node = (*PExpr)(&child)
			case "s":
				e.Node = (*SExpr)(&child)
			case "sub":
				e.Node = (*SubExpr)(&child)
			case "has":
				e.Node = (*HasExpr)(&child)
			case "gg":
				e.Node = (*GGExpr)(&child)
			case "ll":
				e.Node = (*LLExpr)(&child)
			case "gge":
				e.Node = (*GGEExpr)(&child)
			case "lle":
				e.Node = (*LLEExpr)(&child)
			case "cnt":
				e.Node = (*CNTExpr)(&child)
			}
		case "not":
			var child Expr
			if err := json.Unmarshal(val, &child); err != nil {
				return err
			}
			e.Node = (*NOTExpr)(&child)

		case "and", "or":
			var child List
			if err := json.Unmarshal(val, &child); err != nil {
				return err
			}

			if len(child) == 0 {
				return fmt.Errorf("Empty `%s' expression", key)
			}

			switch key {
			case "and":
				e.Node = ANDExpr(child)
			case "or":
				e.Node = ORExpr(child)
			}

		default:
			return fmt.Errorf("Unknown operator `%s'", key)
		}
	}

	return nil
}

func (e *Expr) HasColumn(name string) bool {
	return !nodeWalk(e.Node, func(n Node, key string, value interface{}) bool { return key != name })
}

func (d *DriverParams) pos(i int) string {
	if fn := d.Pfn; fn != nil {
		return fn(i)
	}

	return DefaultPositional(i)
}

func (d *DriverParams) textType() string {
	if t := d.TextType; t != "" {
		return t
	}

	return DefaultTextType
}

func ColumnExpr(col string, columns Columns) (string, error) {
	if columns == nil {
		return col, nil
	}

	c, ok := columns[col]
	if !ok || c == nil {
		return "", fmt.Errorf("Unknown column `%s'", col)
	}

	return c.Expr(), nil
}

func (k *KeyVal) binary(op string, index *int, p *DriverParams, columns Columns) (sql string, args []interface{}, err error) {
	col, err := ColumnExpr(k.Key, columns)
	if err != nil {
		return "", nil, err
	}
	(*index)++
	return col + " " + op + " " + p.pos(*index), []interface{}{k.Value}, nil
}

func (l List) sql(index *int, p *DriverParams, columns Columns) (exprs []string, args []interface{}, err error) {
	exprs = make([]string, len(l))
	args = []interface{}{}

	for i, v := range l {
		e, a, err := v.SQL(index, p, columns)
		if err != nil {
			return nil, nil, err
		}
		exprs[i] = "(" + e + ")"
		args = append(args, a...)
	}

	return exprs, args, nil
}

type ANDExpr List

func (a ANDExpr) SQL(index *int, p *DriverParams, columns Columns) (sql string, args []interface{}, err error) {
	exprs, args, err := List(a).sql(index, p, columns)
	if err != nil {
		return "", nil, err
	}
	return strings.Join(exprs, " AND "), args, nil
}

type ORExpr List

func (or ORExpr) SQL(index *int, p *DriverParams, columns Columns) (sql string, args []interface{}, err error) {
	exprs, args, err := List(or).sql(index, p, columns)
	if err != nil {
		return "", nil, err
	}
	return strings.Join(exprs, " OR "), args, nil
}

type NOTExpr Expr

func (n *NOTExpr) SQL(index *int, p *DriverParams, columns Columns) (sql string, args []interface{}, err error) {
	expr, args, err := (*Expr)(n).SQL(index, p, columns)
	if err != nil {
		return "", nil, err
	}
	return "NOT (" + expr + ")", args, nil
}

type EQExpr KeyVal

func (e *EQExpr) SQL(index *int, p *DriverParams, columns Columns) (sql string, args []interface{}, err error) {
	return (*KeyVal)(e).binary("=", index, p, columns)
}

type NEExpr KeyVal

func (e *NEExpr) SQL(index *int, p *DriverParams, columns Columns) (sql string, args []interface{}, err error) {
	return (*KeyVal)(e).binary("<>", index, p, columns)
}

type LTExpr KeyVal

func (e *LTExpr) SQL(index *int, p *DriverParams, columns Columns) (sql string, args []interface{}, err error) {
	return (*KeyVal)(e).binary("<", index, p, columns)
}

type GTExpr KeyVal

func (e *GTExpr) SQL(index *int, p *DriverParams, columns Columns) (sql string, args []interface{}, err error) {
	return (*KeyVal)(e).binary(">", index, p, columns)
}

type LEExpr KeyVal

func (e *LEExpr) SQL(index *int, p *DriverParams, columns Columns) (sql string, args []interface{}, err error) {
	return (*KeyVal)(e).binary("<=", index, p, columns)
}

type GEExpr KeyVal

func (e *GEExpr) SQL(index *int, p *DriverParams, columns Columns) (sql string, args []interface{}, err error) {
	return (*KeyVal)(e).binary(">=", index, p, columns)
}

type GGExpr KeyVal

func (e *GGExpr) SQL(index *int, p *DriverParams, columns Columns) (sql string, args []interface{}, err error) {
	return (*KeyVal)(e).binary(">>", index, p, columns)
}

type LLExpr KeyVal

func (e *LLExpr) SQL(index *int, p *DriverParams, columns Columns) (sql string, args []interface{}, err error) {
	return (*KeyVal)(e).binary("<<", index, p, columns)
}

type GGEExpr KeyVal

func (e *GGEExpr) SQL(index *int, p *DriverParams, columns Columns) (sql string, args []interface{}, err error) {
	return (*KeyVal)(e).binary(">>=", index, p, columns)
}

type LLEExpr KeyVal

func (e *LLEExpr) SQL(index *int, p *DriverParams, columns Columns) (sql string, args []interface{}, err error) {
	return (*KeyVal)(e).binary("<<=", index, p, columns)
}

type CNTExpr KeyVal

func (e *CNTExpr) SQL(index *int, p *DriverParams, columns Columns) (sql string, args []interface{}, err error) {
	return (*KeyVal)(e).binary("&&", index, p, columns)
}

type ReExpr KeyVal

func (e *ReExpr) SQL(index *int, p *DriverParams, columns Columns) (sql string, args []interface{}, err error) {
	return (*KeyVal)(e).binary("~", index, p, columns)
}

type LExpr KeyVal

func (e *LExpr) SQL(index *int, p *DriverParams, columns Columns) (sql string, args []interface{}, err error) {
	return (*KeyVal)(e).binary("LIKE", index, p, columns)
}

type PExpr KeyVal

func (e *PExpr) SQL(index *int, p *DriverParams, columns Columns) (sql string, args []interface{}, err error) {
	col, err := ColumnExpr(e.Key, columns)
	if err != nil {
		return "", nil, err
	}
	(*index)++
	return col + " LIKE CONCAT(CAST(" + p.pos(*index) + " AS " + p.textType() + "), '%')", []interface{}{e.Value}, nil
}

type SExpr KeyVal

func (e *SExpr) SQL(index *int, p *DriverParams, columns Columns) (sql string, args []interface{}, err error) {
	col, err := ColumnExpr(e.Key, columns)
	if err != nil {
		return "", nil, err
	}
	(*index)++
	return col + " LIKE CONCAT('%', CAST(" + p.pos(*index) + " AS " + p.textType() + "))", []interface{}{e.Value}, nil
}

type SubExpr KeyVal

func (e *SubExpr) SQL(index *int, p *DriverParams, columns Columns) (sql string, args []interface{}, err error) {
	col, err := ColumnExpr(e.Key, columns)
	if err != nil {
		return "", nil, err
	}
	(*index)++
	return col + " LIKE CONCAT('%', CAST(" + p.pos(*index) + " AS " + p.textType() + "), '%')", []interface{}{e.Value}, nil
}

type HasExpr KeyVal

func (e *HasExpr) SQL(index *int, p *DriverParams, columns Columns) (sql string, args []interface{}, err error) {
	col, err := ColumnExpr(e.Key, columns)
	if err != nil {
		return "", nil, err
	}
	(*index)++
	return "(" + col + " IS NOT NULL AND " + p.pos(*index) + " = ANY(" + col + "))", []interface{}{e.Value}, nil
}
