package biznetgio

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
)

// ID is a numeric identifier. The API sends some as numbers and others as
// numeric strings ("account_id": "1234"); ID reads both.
type ID int64

func (id *ID) UnmarshalJSON(b []byte) error {
	b = bytes.TrimSpace(b)
	if string(b) == "null" || string(b) == `""` {
		*id = 0
		return nil
	}
	if len(b) > 0 && b[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return fmt.Errorf("biznetgio: id %q isn't a number", s)
		}
		*id = ID(n)
		return nil
	}
	var n int64
	if err := json.Unmarshal(b, &n); err != nil {
		return err
	}
	*id = ID(n)
	return nil
}

// String is the ID as the API's paths take it.
func (id ID) String() string { return strconv.FormatInt(int64(id), 10) }
