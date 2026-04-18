package output

import (
	"encoding/json"
	"io"
	"strings"

	"github.com/spf13/cobra"
)

// Options holds the parsed output-mode flags for a command.
// Attach it to a command with AddFlags, then call Write in RunE.
type Options struct {
	JSON     bool
	JQ       string
	Template string
	Plain    bool
}

// IsJSON returns true when the active mode produces JSON on stdout
// (--json, --jq, or --template all qualify).
func (o *Options) IsJSON() bool {
	return o.JSON || o.JQ != "" || o.Template != ""
}

// AddFlags wires --json, --jq, --template, and --plain to cmd and appends a
// "JSON Fields:" section to cmd.Long derived from the json tags of example.
// example should be a pointer to (or slice of) the struct used for JSON output.
func AddFlags(cmd *cobra.Command, opts *Options, example interface{}) {
	cmd.Flags().BoolVarP(&opts.JSON, "json", "j", false, "Output as JSON")
	cmd.Flags().StringVar(&opts.JQ, "jq", "", "Filter JSON output with a jq `expression`")
	cmd.Flags().StringVar(&opts.Template, "template", "", "Format output using a Go `template`")
	cmd.Flags().BoolVar(&opts.Plain, "plain", false, "Plain text output (tab-separated, no color)")

	if example != nil {
		if fields := JSONFields(example); len(fields) > 0 {
			cmd.Long += "\n\nJSON Fields:\n  " + strings.Join(fields, ", ")
		}
	}
}

// Write serializes data to w according to the active output mode.
// Returns nil without writing when the mode is human-readable (default or
// --plain), so the caller can handle formatting itself.
func (o *Options) Write(w io.Writer, data interface{}) error {
	switch {
	case o.JQ != "":
		return writeJQ(w, o.JQ, data)
	case o.Template != "":
		return writeTemplate(w, o.Template, data)
	case o.JSON:
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(data)
	default:
		return nil
	}
}
