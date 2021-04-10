package queue

import (
	"fmt"
	"reflect"
)

type ValueWriter interface {
	WriteValue(ptr, value interface{}) error
}

type ValueWriterFunc func(ptr, value interface{}) error

func (v ValueWriterFunc) WriteValue(ptr, value interface{}) error {
	return v(ptr, value)
}

var defaultVW ValueWriterFunc = func(ptr, value interface{}) error {
	vv := reflect.ValueOf(value)
	ptrv := reflect.ValueOf(ptr)
	if ptrv.Kind() != reflect.Ptr && vv.Kind() != reflect.Ptr && ptrv.Type() != reflect.PtrTo(vv.Type()) {
		return fmt.Errorf("incompatible values")
	}
	if vv.Kind() == reflect.Ptr {
		vv = vv.Elem()
	}
	ptrv.Elem().Set(vv)
	return nil
}
