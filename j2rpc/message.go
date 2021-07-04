package j2rpc

import (
	"net/http"
	"strings"

	"github.com/atcharles/gof/v2/json"
)

//AbortWriteHeader ...
func AbortWriteHeader(w http.ResponseWriter, code int) {
	w.WriteHeader(code)
	w.Header().Set("Status-Written", "1")
}

//writeResponse ...
func writeResponse(w http.ResponseWriter, id json.RawMessage, rst ...interface{}) {
	if len(w.Header().Get("Status-Written")) != 0 {
		return
	}

	response := NewResponse(id, rst...)
	bts, err := json.Marshal(response)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(bts)
}

//NewResponse ...
func NewResponse(id json.RawMessage, rst ...interface{}) *RPCMessage {
	var (
		result json.RawMessage
		err    *Error
	)
	if len(rst) > 0 {
		_rst := rst[0]
		switch _d1a := _rst.(type) {
		case json.RawMessage:
			result = _d1a
		case *Error:
			err = _d1a
		case error:
			err = NewError(ErrServer, _d1a.Error())
		default:
			panic("types error")
		}
	}
	if len(id) == 0 {
		id = []byte("1")
	}
	return &RPCMessage{ID: id, Version: vsn, Result: result, Error: err}
}

// RPCMessage A value of this type can a JSON-RPC request, notification, successful response or
//error response. Which one it is depends on the fields.
type RPCMessage struct {
	ID      json.RawMessage `json:"id,omitempty"`
	Version string          `json:"jsonrpc,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *Error          `json:"error,omitempty"`
}

//namespace ...returns the service's name
func (r *RPCMessage) methods() ([]string, error) {
	elem := strings.SplitN(r.Method, splitMethodSeparator, 2)
	if len(elem) != 2 {
		return nil, NewError(ErrNoMethod, "wrong method")
	}
	return elem, nil
}

func (r *RPCMessage) hasValidID() bool { return len(r.ID) > 0 && r.ID[0] != '{' && r.ID[0] != '[' }
