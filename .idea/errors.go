package main

import (
	"errors"
	"fmt"
)

type QueryErrorCause int

const (
	NetworkError QueryErrorCause = iota
	UserError
	UnknownError
)

func (q QueryErrorCause) String() string {
	return [...]string{"NetworkError", "UserError", "UnknownError"}[q]
}

var (
	ErrNetwork = errors.New("socket not found")
	ErrUser    = errors.New("user error")
	ErrUnk     = errors.New("unknown error")
)

type QueryError struct {
	Message string
	Cause   QueryErrorCause
	// We are also adding an error here to allow error wrapping.
	Err error
}

func (e QueryError) Error() string {
	return fmt.Sprintf("Failed to execute the query, %s", e.Message)
}

func (e QueryError) Unwrap() error {
	return e.Err
}

func executeQuery() (int, *QueryError) {
	return 0, &QueryError{
		Message: "network error",
		Cause:   NetworkError,
		Err:     ErrNetwork,
	}
}

func main() {
	_, err := executeQuery()
	if err != nil {
		if errors.Is(err, ErrNetwork) {
			//
		} else if errors.Is(err, ErrUser) {
			//
		} else {
			//
		}
	}

	var queryErr *QueryError
	if errors.As(err, &queryErr) {
		fmt.Printf("QueryError details: cause=%v,msg=%s\n", queryErr.Cause, queryErr.Message)

	}
	otherErr := errors.New("timeout")
	combinedErr := errors.Join(err, otherErr)
	if errors.Is(combinedErr, ErrNetwork) {

	}
}
