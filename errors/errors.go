package errors

import (
	"errors"
	"net/http"
)

type Error struct {
	error
	Code Code
}

func Wrap(e error, code Code) *Error {
	return &Error{e, code}
}

type Response struct {
	Error string `json:"error,omitempty"`
	Code  Code   `json:"code,omitempty"`
}

func (r *Response) HTTPStatus() int {
	return r.Code.HTTPStatus()
}

func ErrorResponse(err error) *Response {
	var code Code

	switch e := err.(type) {
	case *Error:
		code = e.Code
	case Error:
		code = e.Code
	default:
		code = CodeUnknown
	}

	return &Response{
		Error: err.Error(),
		Code:  code,
	}
}

type Code string

func (c Code) HTTPStatus() int {
	if s, ok := httpStatus[c]; ok {
		return s
	}

	return http.StatusInternalServerError
}

const (
	CodeUnknown          Code = "unknown"
	CodeUserNotFound     Code = "user_not_found"
	CodeResourceNotFound Code = "resource_not_found"
	CodeEmailInUse       Code = "email_in_use"
	CodePatchFormat      Code = "patch_format"
	CodeRoleExists       Code = "role_exists"
	CodeTokenExpired     Code = "token_expired"
	CodeBadRequest       Code = "bad_request"
	CodeQuerySyntax      Code = "query_syntax"
	CodeForbidden        Code = "forbidden"
	CodeEmailFmt         Code = "email_format"
	CodeUnauthorized     Code = "unauthorized"
	CodeEmailEmpty       Code = "empty_email"
	CodePasswordEmpty    Code = "empty_password"
	CodeTokenEmpty       Code = "empty_token"
	CodeInvalidToken     Code = "invalid_token"
	CodeTokenFmt         Code = "nvalid_token_fmt"
	CodeAudience         Code = "invalid_audience"
	CodeEmailNotVerified Code = "email_not_verified"
)

var httpStatus = map[Code]int{
	CodeUnknown:          http.StatusInternalServerError,
	CodeUserNotFound:     http.StatusNotFound,
	CodeResourceNotFound: http.StatusNotFound,
	CodeEmailInUse:       http.StatusConflict,
	CodeRoleExists:       http.StatusConflict,
	CodePatchFormat:      http.StatusBadRequest,
	CodeTokenExpired:     http.StatusBadRequest,
	CodeBadRequest:       http.StatusBadRequest,
	CodeQuerySyntax:      http.StatusBadRequest,
	CodeForbidden:        http.StatusForbidden,
	CodeEmailFmt:         http.StatusBadRequest,
	CodeUnauthorized:     http.StatusUnauthorized,
	CodeEmailEmpty:       http.StatusBadRequest,
	CodePasswordEmpty:    http.StatusBadRequest,
	CodeTokenEmpty:       http.StatusBadRequest,
	CodeInvalidToken:     http.StatusUnauthorized,
	CodeAudience:         http.StatusForbidden,
	CodeTokenFmt:         http.StatusBadRequest,
	CodeEmailNotVerified: http.StatusForbidden,
}

// Some predefined errors

var (
	ErrUserNotFound     = &Error{errors.New("User not found"), CodeUserNotFound}
	ErrResourceNotFound = &Error{errors.New("Resource not found"), CodeResourceNotFound}
	ErrEmailInUse       = &Error{errors.New("Email is in use"), CodeEmailInUse}
	ErrPatchValue       = &Error{errors.New("Patch value is missed"), CodePatchFormat}
	ErrRoleExists       = &Error{errors.New("Role exists"), CodeRoleExists}
	ErrTokenExpired     = &Error{errors.New("Token is expired"), CodeTokenExpired}
	ErrEmailFmt         = &Error{errors.New("Invalid email format"), CodeEmailFmt}
	ErrEmailEmpty       = &Error{errors.New("Email value is empty"), CodeEmailEmpty}
	ErrPasswordEmpty    = &Error{errors.New("Password value is empty"), CodePasswordEmpty}
	ErrTokenEmpty       = &Error{errors.New("Token value is empty"), CodeTokenEmpty}
	ErrInvalidToken     = &Error{errors.New("Invalid token"), CodeInvalidToken}
	ErrAudience         = &Error{errors.New("Invalid token audience"), CodeAudience}
	ErrEmailNotVerified = &Error{errors.New("Email is not verified"), CodeEmailNotVerified}
)
