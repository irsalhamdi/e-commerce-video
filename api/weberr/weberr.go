package weberr

type Opt func(error) error

func Wrap(err error, opts ...Opt) error {
	for _, opt := range opts {
		err = opt(err)
	}
	return err
}

func WithResponse(body interface{}, status int) Opt {
	return func(err error) error {
		return &responseError{error: err, body: body, status: status}
	}
}

func WithFields(fields map[string]interface{}) Opt {
	return func(err error) error {
		return &fieldsError{error: err, fields: fields}
	}
}
