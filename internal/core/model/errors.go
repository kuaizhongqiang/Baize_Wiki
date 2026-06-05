package model

import (
	"errors"
	"fmt"
)

// Domain sentinel errors.
var (
	ErrWikiNotFound    = errors.New("wiki not found")
	ErrPageNotFound    = errors.New("page not found")
	ErrSourceNotFound  = errors.New("source not found")
	ErrInvalidConfig   = errors.New("invalid config")
	ErrScanFailed      = errors.New("scan failed")
	ErrGenerateFailed  = errors.New("generate failed")
	ErrEmptySource     = errors.New("empty source")
)

// Error represents a structured domain error with machine-readable code.
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`
	Err     error  `json:"-"`
}

func (e *Error) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Code, e.Message, e.Detail)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *Error) Unwrap() error {
	return e.Err
}
