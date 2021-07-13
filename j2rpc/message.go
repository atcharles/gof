package j2rpc

import (
	"net/http"
	"strings"

	"github.com/atcharles/gof/v2/json"
)

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

//writeResponse ...
func (r *RPCMessage) writeResponse(w http.ResponseWriter) {
	if len(w.Header().Get("Status-Written")) != 0 {
		return
	}

	bts, err := json.Marshal(r)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	AbortWriteHeader(w, http.StatusOK)
	n, err := w.Write(bts)
	_, _ = n, err
}

//output ...
func (r *RPCMessage) output() *RPCMessage {
	if len(r.ID) == 0 {
		r.ID = []byte{'1'}
	}
	r.Version = vsn
	r.Method = ""
	r.Params = nil
	return r
}

//setError ...
func (r *RPCMessage) setError(err error) *RPCMessage {
	if err == nil {
		return r
	}
	var e *Error
	switch _d1a := err.(type) {
	case *Error:
		e = _d1a
	case error:
		e = NewError(ErrServer, _d1a.Error())
	default:
		panic("types error")
	}
	r.Error = e
	return r
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
