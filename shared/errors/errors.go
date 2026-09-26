package errors

import "errors"

var (
	ErrNotFound       = errors.New("resource not found")
	ErrAlreadyExists  = errors.New("resource already exists")
	ErrUnauthorized   = errors.New("unauthorized access")
	ErrInvalidInput   = errors.New("invalid request input")
	ErrConflict       = errors.New("resource conflict state")
	ErrInternalServer = errors.New("internal server error")
	ErrPaymentFailed  = errors.New("payment processing failed")
)
