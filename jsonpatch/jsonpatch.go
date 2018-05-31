package jsonpatch

import (
	"errors"
	"fmt"
	"github.com/lib/pq"
	"strings"
)

type Op struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}

type Patch []*Op

type argIndex int

func (a *argIndex) Next() string {
	(*a)++
	return fmt.Sprintf("$%d", *a)
}

func (o *Op) Column() (string, error) {
	if o.Path != "" && o.Path[0] == '/' && strings.IndexByte(o.Path[1:], '/') < 0 {
		return o.Path[1:], nil
	}
	return "", fmt.Errorf("Incorrect path in JSON patch data: `%s'", o.Path)
}

func errCol(col string) error {
	return fmt.Errorf("Invalid column name `%s'", col)
}

func (p Patch) ValidateFunc(f func(string) bool) error {
	for _, op := range p {
		if op.Op != "replace" {
			return fmt.Errorf("Unknown JSON patch op: `%s'", op.Op)
		}

		col, err := op.Column()
		if err != nil {
			return err
		}

		if f != nil && !f(col) {
			return errCol(col)
		}
	}

	return nil
}

type UpdateOptions struct {
	Table          string
	IDColumn       string
	ID             interface{}
	ReturnUpdated  bool
	ValidateColumn func(string) bool
}

func (p Patch) UpdateStmt(o *UpdateOptions) (string, []interface{}, error) {
	if len(p) == 0 {
		return "", nil, errors.New("Empty patch data")
	}

	if err := p.ValidateFunc(o.ValidateColumn); err != nil {
		return "", nil, err
	}

	expr := "UPDATE " + pq.QuoteIdentifier(o.Table) + " SET "

	arg := make([]interface{}, len(p)+1)
	var idx argIndex

	for i, op := range p {
		if i != 0 {
			expr += ", "
		}

		col, err := op.Column()
		if err != nil {
			return "", nil, err
		}

		expr += pq.QuoteIdentifier(col) + " = " + idx.Next()
		arg[i] = op.Value
	}

	expr += " WHERE " + pq.QuoteIdentifier(o.IDColumn) + " = " + idx.Next()
	arg[len(p)] = o.ID

	if o.ReturnUpdated {
		expr += " RETURNING *"
	}

	return expr, arg, nil
}
