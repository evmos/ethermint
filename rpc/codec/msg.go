package codec

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/log"

	"github.com/evmos/ethermint/rpc/types"
)

// JSONWriter can write JSON messages to its underlying connection.
// Implementations must be safe for concurrent use.
type JSONWriter interface {
	WriteJSON(context.Context, interface{}) error
	// Closed returns a channel which is closed when the connection is closed.
	Closed() <-chan interface{}
	// RemoteAddr returns the peer address of the connection.
	RemoteAddr() string
}

var null = json.RawMessage("null")

type SubscriptionResult struct {
	ID     string          `json:"subscription"`
	Result json.RawMessage `json:"result,omitempty"`
}

// A value of this type can a JSON-RPC request, notification, successful response or
// error response. Which one it is depends on the fields.
type JsonrpcMessage struct {
	Version string          `json:"jsonrpc,omitempty"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Error   *JSONError      `json:"error,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
}

func (msg *JsonrpcMessage) IsNotification() bool {
	return msg.HasValidVersion() && msg.ID == nil && msg.Method != ""
}

func (msg *JsonrpcMessage) IsCall() bool {
	return msg.HasValidVersion() && msg.HasValidID() && msg.Method != ""
}

func (msg *JsonrpcMessage) IsResponse() bool {
	return msg.HasValidVersion() && msg.HasValidID() && msg.Method == "" && msg.Params == nil && (msg.Result != nil || msg.Error != nil)
}

func (msg *JsonrpcMessage) HasValidID() bool {
	return len(msg.ID) > 0 && msg.ID[0] != '{' && msg.ID[0] != '['
}

func (msg *JsonrpcMessage) HasValidVersion() bool {
	return msg.Version == types.Vsn
}

func (msg *JsonrpcMessage) IsSubscribe() bool {
	return strings.HasSuffix(msg.Method, types.SubscribeMethodSuffix)
}

func (msg *JsonrpcMessage) IsUnsubscribe() bool {
	return strings.HasSuffix(msg.Method, types.UnsubscribeMethodSuffix)
}

func (msg *JsonrpcMessage) Namespace() string {
	elem := strings.SplitN(msg.Method, types.ServiceMethodSeparator, 2)
	return elem[0]
}

func (msg *JsonrpcMessage) String() string {
	b, err := json.Marshal(msg)
	if err != nil {
		log.Trace("marshaling jsonrpc message failed", "err", err)
	}
	return string(b)
}

func (msg *JsonrpcMessage) ErrorResponse(err error) *JsonrpcMessage {
	resp := ErrorMessage(err)
	resp.ID = msg.ID
	return resp
}

func (msg *JsonrpcMessage) Response(result interface{}) *JsonrpcMessage {
	enc, err := json.Marshal(result)
	if err != nil {
		return msg.ErrorResponse(&types.InternalServerError{Code: types.ErrcodeMarshalError, Message: err.Error()})
	}
	return &JsonrpcMessage{Version: types.Vsn, ID: msg.ID, Result: enc}
}

func ErrorMessage(err error) *JsonrpcMessage {
	msg := &JsonrpcMessage{Version: types.Vsn, ID: null, Error: &JSONError{
		Code:    types.ErrcodeDefault,
		Message: err.Error(),
	}}
	ec, ok := err.(types.Error)
	if ok {
		msg.Error.Code = ec.ErrorCode()
	}
	de, ok := err.(types.DataError)
	if ok {
		msg.Error.Data = de.ErrorData()
	}
	return msg
}

type JSONError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func (err *JSONError) Error() string {
	if err.Message == "" {
		return fmt.Sprintf("json-rpc error %d", err.Code)
	}
	return err.Message
}

func (err *JSONError) ErrorCode() int {
	return err.Code
}

func (err *JSONError) ErrorData() interface{} {
	return err.Data
}

// Conn is a subset of the methods of net.Conn which are sufficient for ServerCodec.
type Conn interface {
	io.ReadWriteCloser
	SetWriteDeadline(time.Time) error
}

type deadlineCloser interface {
	io.Closer
	SetWriteDeadline(time.Time) error
}

// ConnRemoteAddr wraps the RemoteAddr operation, which returns a description
// of the peer address of a connection. If a Conn also implements ConnRemoteAddr, this
// description is used in log messages.
type ConnRemoteAddr interface {
	RemoteAddr() string
}

// parseMessage parses raw bytes as a (batch of) JSON-RPC message(s). There are no error
// checks in this function because the raw message has already been syntax-checked when it
// is called. Any non-JSON-RPC messages in the input return the zero value of
// jsonrpcMessage.
func parseMessage(raw json.RawMessage) ([]*JsonrpcMessage, bool) {
	if !IsBatch(raw) {
		msgs := []*JsonrpcMessage{{}}
		err := json.Unmarshal(raw, &msgs[0])
		if err != nil {
			log.Trace("unmarshal json message failed", "err", err)
		}
		return msgs, false
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	_, err := dec.Token()
	if err != nil {
		log.Trace("getting next json token failed", "err", err)
	}
	var msgs []*JsonrpcMessage
	for dec.More() {
		msgs = append(msgs, new(JsonrpcMessage))
		err := dec.Decode(&msgs[len(msgs)-1])
		if err != nil {
			log.Trace("decode msg failed", "err", err)
		}
	}
	return msgs, true
}

// IsBatch returns true when the first non-whitespace characters is '['
func IsBatch(raw json.RawMessage) bool {
	for _, c := range raw {
		// skip insignificant whitespace (http://www.ietf.org/rfc/rfc4627.txt)
		if c == 0x20 || c == 0x09 || c == 0x0a || c == 0x0d {
			continue
		}
		return c == '['
	}
	return false
}

// parsePositionalArguments tries to parse the given args to an array of values with the
// given types. It returns the parsed values or an error when the args could not be
// parsed. Missing optional arguments are returned as reflect.Zero values.
func ParsePositionalArguments(rawArgs json.RawMessage, types []reflect.Type) ([]reflect.Value, error) {
	dec := json.NewDecoder(bytes.NewReader(rawArgs))
	var args []reflect.Value
	tok, err := dec.Token()
	switch {
	case err == io.EOF || tok == nil && err == nil:
		// "params" is optional and may be empty. Also allow "params":null even though it's
		// not in the spec because our own client used to send it.
	case err != nil:
		return nil, err
	case tok == json.Delim('['):
		// Read argument array.
		if args, err = ParseArgumentArray(dec, types); err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("non-array args")
	}
	// Set any missing args to nil.
	for i := len(args); i < len(types); i++ {
		if types[i].Kind() != reflect.Ptr {
			return nil, fmt.Errorf("missing value for required argument %d", i)
		}
		args = append(args, reflect.Zero(types[i]))
	}
	return args, nil
}

func ParseArgumentArray(dec *json.Decoder, types []reflect.Type) ([]reflect.Value, error) {
	args := make([]reflect.Value, 0, len(types))
	for i := 0; dec.More(); i++ {
		if i >= len(types) {
			return args, fmt.Errorf("too many arguments, want at most %d", len(types))
		}
		argval := reflect.New(types[i])
		if err := dec.Decode(argval.Interface()); err != nil {
			return args, fmt.Errorf("invalid argument %d: %v", i, err)
		}
		if argval.IsNil() && types[i].Kind() != reflect.Ptr {
			return args, fmt.Errorf("missing value for required argument %d", i)
		}
		args = append(args, argval.Elem())
	}
	// Read end of args array.
	_, err := dec.Token()
	return args, err
}

// parseSubscriptionName extracts the subscription name from an encoded argument array.
func ParseSubscriptionName(rawArgs json.RawMessage) (string, error) {
	dec := json.NewDecoder(bytes.NewReader(rawArgs))
	if tok, _ := dec.Token(); tok != json.Delim('[') {
		return "", errors.New("non-array args")
	}
	v, _ := dec.Token()
	method, ok := v.(string)
	if !ok {
		return "", errors.New("expected subscription name as first argument")
	}
	return method, nil
}
