package biznetgio

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Int is a whole number. The API sends numbers as JSON numbers or as
// strings ("cores": "1", "price": "59000.00"); Int reads both, rounding
// decimals (prices are whole rupiah). "" and null read as 0.
type Int int64

func (n *Int) UnmarshalJSON(b []byte) error {
	b = bytes.TrimSpace(b)
	if string(b) == "null" {
		*n = 0
		return nil
	}
	s := string(b)
	if len(b) > 0 && b[0] == '"' {
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		s = strings.TrimSpace(s)
		if s == "" {
			*n = 0
			return nil
		}
	}
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		*n = Int(i)
		return nil
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil || math.IsNaN(f) || math.IsInf(f, 0) {
		return fmt.Errorf("biznetgio: %s isn't a number", b)
	}
	*n = Int(math.Round(f))
	return nil
}
