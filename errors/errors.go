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
	CodeUnknown             Code = "unknown"
	CodeTenantNotFound      Code = "tenant_not_found"
	CodeMembershipNotFound  Code = "membership_not_found"
	CodeUserNotFound        Code = "user_not_found"
	CodeResourceNotFound    Code = "resource_not_found"
	CodeEmailInUse          Code = "email_in_use"
	CodeMembershipExists    Code = "membership_exists"
	CodePatchFormat         Code = "patch_format"
	CodeRoleExists          Code = "role_exists"
	CodeTokenExpired        Code = "token_expired"
	CodeBadRequest          Code = "bad_request"
	CodeQuerySyntax         Code = "query_syntax"
	CodeForbidden           Code = "forbidden"
	CodeEmailFmt            Code = "email_format"
	CodeUnauthorized        Code = "unauthorized"
	CodeEmailEmpty          Code = "empty_email"
	CodePasswordEmpty       Code = "empty_password"
	CodeTokenEmpty          Code = "empty_token"
	CodeRolesEmpty          Code = "empty_roles"
	CodeInvalidToken        Code = "invalid_token"
	CodeTokenFmt            Code = "nvalid_token_fmt"
	CodeAudience            Code = "invalid_audience"
	CodeEmailNotVerified    Code = "email_not_verified"
	CodeRoleNotFound        Code = "role_not_found"
	CodePermissionNotFound  Code = "permission_not_found"
	CodeMembershipNotActive Code = "membership_not_active"
	CodeService             Code = "service_account"
	CodeKeyNotFound         Code = "api_key_not_found"
	CodeAddrExists          Code = "address_exists"
)

var httpStatus = map[Code]int{
	CodeUnknown:             http.StatusInternalServerError,
	CodeUserNotFound:        http.StatusNotFound,
	CodeTenantNotFound:      http.StatusNotFound,
	CodeResourceNotFound:    http.StatusNotFound,
	CodeEmailInUse:          http.StatusConflict,
	CodeRoleExists:          http.StatusConflict,
	CodePatchFormat:         http.StatusBadRequest,
	CodeTokenExpired:        http.StatusBadRequest,
	CodeBadRequest:          http.StatusBadRequest,
	CodeQuerySyntax:         http.StatusBadRequest,
	CodeForbidden:           http.StatusForbidden,
	CodeEmailFmt:            http.StatusBadRequest,
	CodeUnauthorized:        http.StatusUnauthorized,
	CodeEmailEmpty:          http.StatusBadRequest,
	CodePasswordEmpty:       http.StatusBadRequest,
	CodeTokenEmpty:          http.StatusBadRequest,
	CodeInvalidToken:        http.StatusUnauthorized,
	CodeAudience:            http.StatusForbidden,
	CodeTokenFmt:            http.StatusBadRequest,
	CodeEmailNotVerified:    http.StatusForbidden,
	CodeMembershipNotActive: http.StatusForbidden,
	CodeRolesEmpty:          http.StatusBadRequest,
	CodeRoleNotFound:        http.StatusNotFound,
	CodePermissionNotFound:  http.StatusNotFound,
	CodeService:             http.StatusBadRequest,
	CodeKeyNotFound:         http.StatusNotFound,
	CodeAddrExists:          http.StatusConflict,
}

// Some predefined errors

var (
	ErrMembershipExisits   = &Error{errors.New("Membership exists"), CodeMembershipExists}
	ErrTenantName          = &Error{errors.New("Name is required"), CodeBadRequest}
	ErrTenantNotFound      = &Error{errors.New("Tenant not found"), CodeTenantNotFound}
	ErrMembershipNotFound  = &Error{errors.New("Membership not found"), CodeMembershipNotFound}
	ErrUserNotFound        = &Error{errors.New("User not found"), CodeUserNotFound}
	ErrResourceNotFound    = &Error{errors.New("Resource not found"), CodeResourceNotFound}
	ErrEmailInUse          = &Error{errors.New("Email is in use"), CodeEmailInUse}
	ErrPatchValue          = &Error{errors.New("Patch value is missed"), CodePatchFormat}
	ErrRoleExists          = &Error{errors.New("Role exists"), CodeRoleExists}
	ErrTokenExpired        = &Error{errors.New("Token is expired"), CodeTokenExpired}
	ErrEmailFmt            = &Error{errors.New("Invalid email format"), CodeEmailFmt}
	ErrEmailEmpty          = &Error{errors.New("Email value is empty"), CodeEmailEmpty}
	ErrPasswordEmpty       = &Error{errors.New("Password value is empty"), CodePasswordEmpty}
	ErrTokenEmpty          = &Error{errors.New("Token value is empty"), CodeTokenEmpty}
	ErrInvalidToken        = &Error{errors.New("Invalid token"), CodeInvalidToken}
	ErrAudience            = &Error{errors.New("Invalid token audience"), CodeAudience}
	ErrEmailNotVerified    = &Error{errors.New("Email is not verified"), CodeEmailNotVerified}
	ErrMembershipNotActive = &Error{errors.New("Membership not active"), CodeMembershipNotActive}
	ErrForbidden           = &Error{errors.New("Forbidden"), CodeForbidden}
	ErrRolesEmpty          = &Error{errors.New("Roles value is empty"), CodeRolesEmpty}
	ErrRoleNotFound        = &Error{errors.New("Role not found"), CodeRoleNotFound}
	ErrPermissionNotFound  = &Error{errors.New("Permission not found"), CodePermissionNotFound}
	ErrService             = &Error{errors.New("Service account"), CodeService}
	ErrKeyNotFound         = &Error{errors.New("Key not found"), CodeKeyNotFound}
	ErrAddrExists          = &Error{errors.New("Address exists"), CodeAddrExists}
	ErrAddrSyntax          = &Error{errors.New("Error parsing address"), CodeBadRequest}
)
