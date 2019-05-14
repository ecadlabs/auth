package jq

import (
	"reflect"
)

var (
	listType   = reflect.TypeOf(List(nil))
	keyValType = reflect.TypeOf((*KeyVal)(nil))
	exprType   = reflect.TypeOf((*Expr)(nil))
)

type walkFn func(n Node, key string, value interface{}) bool

func nodeWalk(n Node, fn walkFn) bool {
	v := reflect.ValueOf(n)
	t := v.Type()

	switch {
	case t.ConvertibleTo(listType):
		// Node itself
		if !fn(n, "", nil) {
			return false
		}

		// Children
		list := v.Convert(listType).Interface().(List)
		for _, vv := range list {
			if !nodeWalk(vv.Node, fn) {
				return false
			}
		}
	case t.ConvertibleTo(keyValType):
		// Node itself
		kv := v.Convert(keyValType).Interface().(*KeyVal)
		return fn(n, kv.Key, kv.Value)

	case t.ConvertibleTo(exprType):
		// Node itself
		if !fn(n, "", nil) {
			return false
		}

		// Child
		e := v.Convert(exprType).Interface().(*Expr)
		return nodeWalk(e.Node, fn)
	}

	return true
}
