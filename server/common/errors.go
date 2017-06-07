package common

import (
	"fmt"
	"strings"

	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
)

var ErrBadRequest = NewClientError("", "Bad request")
var ErrSystemError = NewClientErrorCode("", "System error", 500)
var ErrUserNotFound = NewClientErrorCode("username", "User not found", 404)
var ErrUserDeleted = NewClientError("username", "User deleted")
var ErrUserEmailNotConfirmed = NewClientError("username", "Email is not confirmed")
var ErrNotFound = NewClientErrorCode("", "Not found", 404)
var ErrAuthorizationRequired = NewClientErrorCode("", "Authorization required", 401)
var ErrPermissionDenied = NewClientErrorCode("", "Permission denied", 403)
var ErrUserCodeExpired = NewClientError("", "Confirmation code is expired")
var ErrUserCodeEmpty = NewClientError("", "Confirmation code is not specified")
var ErrUserCodeNotValid = NewClientError("", "Confirmation code is not valid")

type BaseResp struct {
	Errors []*ClientError `json:"errors,omitempty"`
}

type DummyResp struct {
	BaseResp
}

type AppError interface {
	Code() int
	Message() string
	Error() string
}

type CustomError struct {
	err     error
	message string
	code    int
}

type ClientError struct {
	Code_    int    `json:"-"`
	Field_   string `json:"field"`
	Message_ string `json:"message"`
}

func NewClientError(field, message string) *ClientError {
	return &ClientError{
		Code_:    400,
		Field_:   field,
		Message_: message,
	}
}

func NewClientErrorCode(field, message string, code int) *ClientError {
	return &ClientError{
		Code_:    code,
		Field_:   field,
		Message_: message,
	}
}

func (e *ClientError) Code() int {
	return e.Code_
}

func (e *ClientError) Message() string {
	return e.Message_
}

func (e *ClientError) Error() string {
	if e.Field_ != "" {
		return strings.Join([]string{e.Field_, e.Message_}, " - ")
	}
	return e.Message_
}

func NewAppError(err error) AppError {
	code := 500
	if ee, ok := err.(*ClientError); ok {
		code = ee.Code()
	}
	return &CustomError{
		err:  err,
		code: code,
	}
}

func AppErrorf(err error, format string, v ...interface{}) AppError {
	code := 500
	if ee, ok := err.(*ClientError); ok {
		code = ee.Code()
	}
	return &CustomError{
		err:     err,
		message: fmt.Sprintf(format, v...),
		code:    code,
	}
}

var _ AppError = &CustomError{}
var _ AppError = &ClientError{}

func (myerr *CustomError) Code() int {
	return myerr.code
}

func (myerr *CustomError) Message() string {
	return myerr.message
}

func (myerr *CustomError) Error() string {
	return myerr.err.Error()
}

func PrintClientErrors(ctx context.Context, err AppError) []*ClientError {
	res := make([]*ClientError, 0)
	if ae, ok := err.(*CustomError); ok {
		if ee, ok := ae.err.(*ClientError); ok {
			res = append(res, ee)
		} else {
			res = append(res, NewClientError("", ae.Message()))
		}
	} else if ae, ok := err.(*ClientError); ok {
		res = append(res, ae)
	} else {
		log.Warningf(ctx, "Error: %s", err)
		res = append(res, ErrSystemError)
	}
	return res
}
