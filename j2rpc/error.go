package j2rpc

//ErrorCode ... Error codes
type ErrorCode int

//declared ...
const (
	ErrParse          ErrorCode = -32700
	ErrInvalidRequest ErrorCode = -32600
	ErrNoMethod       ErrorCode = -32601
	ErrBadParams      ErrorCode = -32602
	ErrInternal       ErrorCode = -32603
	ErrServer         ErrorCode = -32000
	ErrAuthorization  ErrorCode = 401
)

//Error ... Error codes
type Error struct {
	Code    ErrorCode   `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func (e *Error) Error() string { return e.Message }

//NewError ...
func NewError(code ErrorCode, Msg string, data ...interface{}) *Error {
	ee := &Error{Code: code, Message: Msg}
	if len(data) > 0 {
		ee.Data = data[0]
	}
	return ee
}

//TokenError ...
type TokenError string

func (t TokenError) Error() string { return string(t) }
