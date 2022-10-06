// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package types

import (
	"errors"
	"fmt"
)

var (
	// ErrNotificationsUnsupported is returned when the connection doesn't support notifications
	ErrNotificationsUnsupported = errors.New("notifications not supported")
	// ErrSubscriptionNotFound is returned when the notification for the given id is not found
	ErrSubscriptionNotFound = errors.New("subscription not found")
)

// HTTPError is returned by client operations when the HTTP status code of the
// response is not a 2xx status.
type HTTPError struct {
	StatusCode int
	Status     string
	Body       []byte
}

func (err HTTPError) Error() string {
	if len(err.Body) == 0 {
		return err.Status
	}
	return fmt.Sprintf("%v: %s", err.Status, err.Body)
}

// Error wraps RPC errors, which contain an error code in addition to the message.
type Error interface {
	Error() string  // returns the message
	ErrorCode() int // returns the code
}

// A DataError contains some data in addition to the error message.
type DataError interface {
	Error() string          // returns the message
	ErrorData() interface{} // returns the error data
}

// Error types defined below are the built-in JSON-RPC errors.

var (
	_ Error = new(MethodNotFoundError)
	_ Error = new(SubscriptionNotFoundError)
	_ Error = new(ParseError)
	_ Error = new(InvalidRequestError)
	_ Error = new(InvalidMessageError)
	_ Error = new(InvalidParamsError)
	_ Error = new(InternalServerError)
)

const (
	ErrcodeDefault                  = -32000
	ErrcodeNotificationsUnsupported = -32001
	ErrcodePanic                    = -32603
	ErrcodeMarshalError             = -32603
)

type MethodNotFoundError struct{ Method string }

func (e *MethodNotFoundError) ErrorCode() int { return -32601 }

func (e *MethodNotFoundError) Error() string {
	return fmt.Sprintf("the method %s does not exist/is not available", e.Method)
}

type SubscriptionNotFoundError struct{ Namespace, Subscription string }

func (e *SubscriptionNotFoundError) ErrorCode() int { return -32601 }

func (e *SubscriptionNotFoundError) Error() string {
	return fmt.Sprintf("no %q subscription in %s namespace", e.Subscription, e.Namespace)
}

// Invalid JSON was received by the server.
type ParseError struct{ Message string }

func (e *ParseError) ErrorCode() int { return -32700 }

func (e *ParseError) Error() string { return e.Message }

// received message isn't a valid request
type InvalidRequestError struct{ Message string }

func (e *InvalidRequestError) ErrorCode() int { return -32600 }

func (e *InvalidRequestError) Error() string { return e.Message }

// received message is invalid
type InvalidMessageError struct{ Message string }

func (e *InvalidMessageError) ErrorCode() int { return -32700 }

func (e *InvalidMessageError) Error() string { return e.Message }

// unable to decode supplied params, or an invalid number of parameters
type InvalidParamsError struct{ Message string }

func (e *InvalidParamsError) ErrorCode() int { return -32602 }

func (e *InvalidParamsError) Error() string { return e.Message }

// internalServerError is used for server errors during request processing.
type InternalServerError struct {
	Code    int
	Message string
}

func (e *InternalServerError) ErrorCode() int { return e.Code }

func (e *InternalServerError) Error() string { return e.Message }

type WsHandshakeError struct {
	Err    error
	Status string
}

func (e WsHandshakeError) Error() string {
	s := e.Err.Error()
	if e.Status != "" {
		s += " (HTTP status " + e.Status + ")"
	}
	return s
}
