package weberr

import "errors"

type responder interface {
	Response() (body interface{}, status int)
}

func Response(err error) (body interface{}, status int, ok bool) {
	var re responder
	if errors.As(err, &re) {
		body, code := re.Response()
		return body, code, true
	}
	return nil, 0, false
}

type responseError struct {
	error
	body   interface{}
	status int
}

func (e *responseError) Response() (interface{}, int) {
	return e.body, e.status
}

func (e *responseError) Unwrap() error {
	return e.error
}
