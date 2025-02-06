// Package errors provides error response functionality.
package errors

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"
)

// Error values contain information about error conditions.
type Error struct {
	Code
	Msg    string         `json:"message,omitempty"`
	Proc   string         `json:"procedure,omitempty"`
	Svr    string         `json:"server,omitempty"`
	Time   int64          `json:"time,omitempty"`
	Data   map[string]any `json:"data,omitempty"`
	Err    *Error         `json:"error,omitempty"`
	Errors []*Error       `json:"errors,omitempty"`
	err    error          `json:"-"`
}

// Code values represent specific error codes and status values.
type Code struct {
	Name   string `json:"code,omitempty"`
	Status int    `json:"status,omitempty"`
}

// dataToArgs converts a []any of args into an error data map[string]any.
func argsToData(args []any) map[string]any {
	data := map[string]any{}

	key := ""

	for i, a := range args {
		switch v := a.(type) {
		case string:
			if i%2 == 0 {
				key = v
			} else {
				if key == "" {
					continue
				}

				data[key] = v
				key = ""
			}
		default:
			if i%2 == 0 {
				continue
			} else {
				if key == "" {
					continue
				}

				data[key] = v
				key = ""
			}
		}
	}

	return data
}

// dataToArgs converts an error data map[string]any into a []any.
func dataToArgs(data map[string]any) []any {
	args := []any{}

	for _, v := range data {
		args = append(args, v)
	}

	return args
}

// New creates a new error value.
func New(code Code, message string, args ...any) *Error {
	var data map[string]any

	if len(args) > 0 {
		data = argsToData(args)
	}

	if code.Status < http.StatusOK ||
		code.Status > http.StatusNetworkAuthenticationRequired {
		code.Status = http.StatusInternalServerError
	}

	e := &Error{
		Code: code,
		Msg:  message,
		Time: time.Now().Unix(),
		Data: data,
	}

	caller := 1

	if pc, _, _, ok := runtime.Caller(caller); ok {
		if f := runtime.FuncForPC(pc); f != nil {
			e.Proc = f.Name()[strings.LastIndex(f.Name(), "/")+1:]
		}
	}

	// Work back up the calling stack until a procedure outside this package
	// is found to use for the calling procedure.
	for strings.HasPrefix(e.Proc, "errors") ||
		strings.Contains(e.Proc, "AsError") {
		caller++

		pc, _, _, ok := runtime.Caller(caller)

		if !ok {
			break
		}

		if f := runtime.FuncForPC(pc); f != nil {
			e.Proc = f.Name()[strings.LastIndex(f.Name(), "/")+1:]
		}
	}

	if host, err := os.Hostname(); err == nil {
		e.Svr = host
	}

	return e
}

// Validation returns a new invalid value error for the specified object
// and field for the provided invalid value.
func Validation(obj, field string, value any) error {
	return New(ErrInvalidRequest,
		fmt.Sprintf("invalid %s %s: %v", obj, field, value))
}

// Context creates a new error value wrapping a context error.
func Context(ctx context.Context) *Error {
	if ctx.Err() == nil {
		return nil
	}

	switch ctx.Err() {
	case context.Canceled:
		return Wrap(ctx.Err(), ErrContextCanceled, ctx.Err().Error())
	case context.DeadlineExceeded:
		return Wrap(ctx.Err(), ErrContextTimeout, ctx.Err().Error())
	default:
		return Wrap(ctx.Err(), ErrContext, ctx.Err().Error())
	}
}

// Wrap creates a new error value from an existing error.
func Wrap(err error, code Code, message string, args ...any) *Error {
	e := New(code, message, args...)

	if ev, ok := err.(*Error); ok {
		e.Err = ev
		e.Time = ev.Time
		e.Code = ev.Code

		if message == "" {
			e.Msg = ev.Msg
		}
	} else if err != nil {
		e.Err = &Error{Msg: err.Error(), err: err}
	}

	caller := 1

	if pc, _, _, ok := runtime.Caller(caller); ok {
		if f := runtime.FuncForPC(pc); f != nil {
			e.Proc = f.Name()[strings.LastIndex(f.Name(), "/")+1:]
		}
	}

	// Work back up the calling stack until a procedure outside this package
	// is found to use for the calling procedure.
	for strings.HasPrefix(e.Proc, "errors") ||
		strings.HasPrefix(e.Proc, "logger") ||
		strings.Contains(e.Proc, "AsError") {
		caller++

		pc, _, _, ok := runtime.Caller(caller)

		if !ok {
			break
		}

		if f := runtime.FuncForPC(pc); f != nil {
			e.Proc = f.Name()[strings.LastIndex(f.Name(), "/")+1:]
		}
	}

	return e
}

// As implemented for compatibility with go standard library errors package.
func As(err error, target any) bool {
	return errors.As(err, target)
}

// Is implemented for compatibility with go standard library errors package.
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// Has returns whether an error has a specified error code.
func Has(err error, code Code) bool {
	if e, ok := err.(*Error); ok {
		if e.Name == code.Name && e.Status == code.Status {
			return true
		}
	}

	return false
}

// String returns the Error object as a string.
func (e *Error) String() string {
	str, err := json.Marshal(e)
	if err != nil {
		return ""
	}

	return string(str)
}

// Error returns the Error object formatted as a JSON string.
func (e *Error) Error() string {
	return e.String()
}

// Unwrap returns the error wrapped by this error, nil if no error is wrapped.
func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}

	if e.err != nil {
		return e.err
	}

	return e.Err
}

// Wrap returns an error value wrapping an existing error value.
func (e *Error) Wrap(err error) *Error {
	return Wrap(err, e.Code, e.Msg, e.Data)
}

// Copy returns an exact copy of the value.
func (e *Error) Copy() *Error {
	err := &Error{
		Code: e.Code,
		Msg:  e.Msg,
		Proc: e.Proc,
		Svr:  e.Svr,
		Time: e.Time,
		Err:  e.Err,
		err:  e.err,
	}

	if len(e.Errors) > 0 {
		err.Errors = make([]*Error, len(e.Errors))
		copy(err.Errors, e.Errors)
	}

	if len(e.Data) > 0 {
		err.Data = make(map[string]any, len(e.Data))

		for k, v := range e.Data {
			err.Data[k] = v
		}
	}

	return err
}

// Equal tests for equality between error values.
func (e *Error) Equal(b *Error) bool {
	switch {
	case e == nil || b == nil:
		return false
	case e == b:
		return true
	case e.Code.Name != b.Code.Name || e.Code.Status != b.Code.Status:
		return false
	case e.Msg != b.Msg:
		return false
	case e.Proc != b.Proc:
		return false
	case e.Svr != b.Svr:
		return false
	case e.Time != b.Time:
		return false
	case e.Err == nil && b.Err != nil ||
		e.Err != nil && b.Err == nil:
		return false
	case e.Err != nil && b.Err != nil &&
		e.Err.Error() != b.Err.Error():
		return false
	case len(e.Errors) != len(b.Errors):
		return false
	case len(e.Data) != len(b.Data):
		return false
	}

	for i, err := range e.Errors {
		if err.Error() != b.Errors[i].Error() {
			return false
		}
	}

	for k, v := range e.Data {
		if b.Data[k] != v {
			return false
		}
	}

	return true
}

// ErrorHas returns true if the provided error as a string contains s.
func ErrorHas(err error, s string) bool {
	if err == nil {
		return false
	}

	return strings.Contains(err.Error(), s)
}

// Defined error codes.
const (
	// Non-standard status code for client closing request.
	StatusCanceled = 499
)

var (
	ErrInvalidRequest = Code{
		Name:   "InvalidRequest",
		Status: http.StatusBadRequest,
	}

	ErrInvalidHeader = Code{
		Name:   "InvalidHeader",
		Status: http.StatusBadRequest,
	}

	ErrInvalidParameter = Code{
		Name:   "InvalidParameter",
		Status: http.StatusBadRequest,
	}

	ErrUnauthorized = Code{
		Name:   "Unauthorized",
		Status: http.StatusUnauthorized,
	}

	ErrForbidden = Code{
		Name:   "Forbidden",
		Status: http.StatusForbidden,
	}

	ErrNotFound = Code{
		Name:   "NotFound",
		Status: http.StatusNotFound,
	}

	ErrNotAllowed = Code{
		Name:   "NotAllowed",
		Status: http.StatusMethodNotAllowed,
	}

	ErrConflict = Code{
		Name:   "Conflict",
		Status: http.StatusConflict,
	}

	ErrServer = Code{
		Name:   "Server",
		Status: http.StatusInternalServerError,
	}

	ErrContext = Code{
		Name:   "Context",
		Status: http.StatusInternalServerError,
	}

	ErrContextCanceled = Code{
		Name:   "Canceled",
		Status: StatusCanceled,
	}

	ErrContextTimeout = Code{
		Name:   "Timeout",
		Status: http.StatusInternalServerError,
	}

	ErrLog = Code{
		Name:   "Log",
		Status: http.StatusInternalServerError,
	}

	ErrMetric = Code{
		Name:   "Metric",
		Status: http.StatusInternalServerError,
	}

	ErrTrace = Code{
		Name:   "Trace",
		Status: http.StatusInternalServerError,
	}

	ErrCache = Code{
		Name:   "Cache",
		Status: http.StatusInternalServerError,
	}

	ErrClient = Code{
		Name:   "Client",
		Status: http.StatusInternalServerError,
	}

	ErrInstall = Code{
		Name:   "Install",
		Status: http.StatusInternalServerError,
	}

	ErrConfiguration = Code{
		Name:   "Configuration",
		Status: http.StatusInternalServerError,
	}

	ErrDatabase = Code{
		Name:   "Database",
		Status: http.StatusInternalServerError,
	}

	ErrSearch = Code{
		Name:   "Search",
		Status: http.StatusInternalServerError,
	}

	ErrImport = Code{
		Name:   "Import",
		Status: http.StatusInternalServerError,
	}

	ErrMaintenance = Code{
		Name:   "Maintenance",
		Status: http.StatusServiceUnavailable,
	}

	ErrUnavailable = Code{
		Name:   "Unavailable",
		Status: http.StatusServiceUnavailable,
	}

	ErrUnimplemented = Code{
		Name:   "Unimplemented",
		Status: http.StatusNotImplemented,
	}

	ErrorRateLimit = Code{
		Name:   "RateLimit",
		Status: http.StatusTooManyRequests,
	}
)
