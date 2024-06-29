package weberr

import (
	"net/http"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type RequestError struct {
	Err error
}

func (r *RequestError) Error() string { return r.Err.Error() }

func (e *RequestError) Unwrap() error { return e.Err }

func NewError(err error, msg string, status int, opts ...Opt) error {
	e := &RequestError{Err: err}
	opts = append(opts, WithResponse(
		&ErrorResponse{msg},
		status,
	))

	return Wrap(e, opts...)
}

func NotFound(err error, opts ...Opt) error {
	return NewError(
		err,
		"the resource could not be found",
		http.StatusNotFound,
		opts...,
	)
}

func NotAuthorized(err error, opts ...Opt) error {
	return NewError(
		err,
		"not authorized to access resource",
		http.StatusUnauthorized,
		opts...,
	)
}

func InternalError(err error, opts ...Opt) error {
	return NewError(
		err,
		"the server encountered a problem and could not process your request",
		http.StatusInternalServerError,
		opts...,
	)
}

func BadRequest(err error, opts ...Opt) error {
	return NewError(
		err,
		"bad request",
		http.StatusBadRequest,
		opts...,
	)
}
