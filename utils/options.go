package utils

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

type Options map[string]interface{}

var ErrOptNotFound = errors.New("Option not found")

func (o Options) GetString(name string) (string, error) {
	v, ok := o[name]
	if !ok {
		return "", ErrOptNotFound
	}

	switch vv := v.(type) {
	case string:
		return vv, nil
	default:
		return fmt.Sprintf("%v", v), nil
	}
}

func getInt(v interface{}) int64 {
	val := reflect.ValueOf(v)

	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return val.Int()

	case reflect.Float32, reflect.Float64:
		return int64(val.Float())
	}

	return 0
}

func (o Options) GetInt(name string) (int64, error) {
	v, ok := o[name]
	if !ok {
		return 0, ErrOptNotFound
	}

	if s, ok := v.(string); ok {
		return strconv.ParseInt(s, 0, 64)
	}

	return getInt(v), nil
}

func (o Options) GetBool(name string) (bool, error) {
	v, ok := o[name]
	if !ok {
		return false, ErrOptNotFound
	}

	switch vv := v.(type) {
	case bool:
		return vv, nil
	case string:
		return strconv.ParseBool(vv)
	default:
		if getInt(v) != 0 {
			return true, nil
		}
		return false, nil
	}
}
