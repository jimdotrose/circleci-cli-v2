package output

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/itchyny/gojq"
)

// writeJQ runs expr against data and writes each result value to w.
// String scalars are written without JSON quoting; everything else is
// pretty-printed JSON.
func writeJQ(w io.Writer, expr string, data interface{}) error {
	query, err := gojq.Parse(expr)
	if err != nil {
		return fmt.Errorf("invalid --jq expression: %w", err)
	}

	// Round-trip through JSON so gojq gets plain map/slice/scalar values.
	raw, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("jq: marshaling input: %w", err)
	}
	var input interface{}
	if err := json.Unmarshal(raw, &input); err != nil {
		return fmt.Errorf("jq: unmarshaling input: %w", err)
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")

	iter := query.Run(input)
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			return fmt.Errorf("jq: %w", err)
		}
		switch val := v.(type) {
		case string:
			fmt.Fprintln(w, val)
		case nil:
			fmt.Fprintln(w, "null")
		default:
			if err := enc.Encode(val); err != nil {
				return err
			}
		}
	}
	return nil
}
