package weberr

import "errors"

type fielder interface {
	Fields() map[string]interface{}
}

func Fields(err error) (fields map[string]interface{}, ok bool) {
	var fe fielder
	if errors.As(err, &fe) {
		return fe.Fields(), true
	}
	return nil, false
}

type fieldsError struct {
	error
	fields map[string]interface{}
}

func (e *fieldsError) Fields() map[string]interface{} { return e.fields }

func (e *fieldsError) Unwrap() error { return e.error }
