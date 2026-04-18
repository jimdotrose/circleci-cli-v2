package output

import (
	"encoding/json"
	"fmt"
	"io"
	"text/template"
)

// writeTemplate renders tmplStr against data and writes the result to w.
// data is round-tripped through JSON before template execution so the
// template receives plain map/slice/scalar values regardless of the
// original Go type.
func writeTemplate(w io.Writer, tmplStr string, data interface{}) error {
	tmpl, err := template.New("").Parse(tmplStr)
	if err != nil {
		return fmt.Errorf("invalid --template: %w", err)
	}

	raw, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("template: marshaling input: %w", err)
	}
	var input interface{}
	if err := json.Unmarshal(raw, &input); err != nil {
		return fmt.Errorf("template: unmarshaling input: %w", err)
	}

	if err := tmpl.Execute(w, input); err != nil {
		return fmt.Errorf("template: executing: %w", err)
	}
	return nil
}
