package g2util

import (
	"encoding/base64"

	"github.com/andeya/goutil"

	"github.com/atcharles/gof/v2/json"
)

// JSONDump ...
func JSONDump(val interface{}, args ...interface{}) string {
	var indent bool
	if len(args) > 0 {
		indent, _ = args[0].(bool)
	}
	if indent {
		return goutil.BytesToString(JsMarshalIndent(val))
	}
	return goutil.BytesToString(JsMarshal(val))
}

// JsMarshal ...
func JsMarshal(val interface{}) (bts []byte) { bts, _ = json.Marshal(val); return }

// JsMarshalIndent ...
func JsMarshalIndent(val interface{}) (bts []byte) {
	bts, _ = json.MarshalIndent(val, "", "\t")
	return
}

// JSONUnmarshalFromBase64 ...
func JSONUnmarshalFromBase64(data []byte, val interface{}) error {
	enc := base64.StdEncoding
	dbuf := make([]byte, enc.DecodedLen(len(data)))
	n, err := enc.Decode(dbuf, data)
	if err != nil {
		return err
	}
	bts := dbuf[:n]
	return json.Unmarshal(bts, val)
}

// JSONMarshalToBase64 ...
func JSONMarshalToBase64(val interface{}) ([]byte, error) {
	bts, err := json.Marshal(val)
	if err != nil {
		return bts, err
	}
	enc := base64.StdEncoding
	buf := make([]byte, enc.EncodedLen(len(bts)))
	enc.Encode(buf, bts)
	return buf, err
}
