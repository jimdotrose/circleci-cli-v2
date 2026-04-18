package output_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/CircleCI-Public/circleci-cli/pkg/output"
)

type testItem struct {
	ID    string `json:"id"`
	State string `json:"state"`
	Count int    `json:"count"`
}

func TestJSONFields(t *testing.T) {
	fields := output.JSONFields(new(testItem))
	want := []string{"id", "state", "count"}
	if len(fields) != len(want) {
		t.Fatalf("JSONFields = %v; want %v", fields, want)
	}
	for i, f := range fields {
		if f != want[i] {
			t.Errorf("field[%d] = %q; want %q", i, f, want[i])
		}
	}
}

func TestJSONFields_slice(t *testing.T) {
	// Slice of structs should give same fields as the element type.
	fields := output.JSONFields([]testItem{})
	if len(fields) != 3 {
		t.Fatalf("JSONFields(slice) = %v; want 3 fields", fields)
	}
}

func TestWrite_JSON(t *testing.T) {
	opts := &output.Options{JSON: true}
	var buf bytes.Buffer
	items := []testItem{{ID: "abc", State: "running", Count: 1}}
	if err := opts.Write(&buf, items); err != nil {
		t.Fatalf("Write JSON: %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, `"id": "abc"`) {
		t.Errorf("JSON output missing id field: %s", got)
	}
}

func TestWrite_JQ(t *testing.T) {
	opts := &output.Options{JQ: ".[0].id"}
	var buf bytes.Buffer
	items := []testItem{{ID: "abc", State: "running", Count: 1}}
	if err := opts.Write(&buf, items); err != nil {
		t.Fatalf("Write JQ: %v", err)
	}
	if strings.TrimSpace(buf.String()) != "abc" {
		t.Errorf("JQ output = %q; want abc", buf.String())
	}
}

func TestWrite_JQ_count(t *testing.T) {
	opts := &output.Options{JQ: "length"}
	var buf bytes.Buffer
	items := []testItem{{ID: "a"}, {ID: "b"}, {ID: "c"}}
	if err := opts.Write(&buf, items); err != nil {
		t.Fatalf("Write JQ length: %v", err)
	}
	if strings.TrimSpace(buf.String()) != "3" {
		t.Errorf("JQ length = %q; want 3", buf.String())
	}
}

func TestWrite_Template(t *testing.T) {
	opts := &output.Options{Template: `{{range .}}{{.id}}{{"\n"}}{{end}}`}
	var buf bytes.Buffer
	items := []testItem{{ID: "abc"}, {ID: "def"}}
	if err := opts.Write(&buf, items); err != nil {
		t.Fatalf("Write template: %v", err)
	}
	if !strings.Contains(buf.String(), "abc") || !strings.Contains(buf.String(), "def") {
		t.Errorf("template output = %q; want abc and def", buf.String())
	}
}

func TestWrite_default_noop(t *testing.T) {
	opts := &output.Options{}
	var buf bytes.Buffer
	if err := opts.Write(&buf, []testItem{{ID: "x"}}); err != nil {
		t.Fatalf("Write default: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("default mode wrote %q; want empty (caller handles formatting)", buf.String())
	}
}

func TestWrite_invalidJQ(t *testing.T) {
	opts := &output.Options{JQ: ".[[[invalid"}
	var buf bytes.Buffer
	if err := opts.Write(&buf, []testItem{}); err == nil {
		t.Error("invalid jq should return error")
	}
}

func TestWrite_invalidTemplate(t *testing.T) {
	opts := &output.Options{Template: "{{.unclosed"}
	var buf bytes.Buffer
	if err := opts.Write(&buf, []testItem{}); err == nil {
		t.Error("invalid template should return error")
	}
}

func TestIsJSON(t *testing.T) {
	cases := []struct {
		opts output.Options
		want bool
	}{
		{output.Options{JSON: true}, true},
		{output.Options{JQ: ".foo"}, true},
		{output.Options{Template: "{{.}}"}, true},
		{output.Options{Plain: true}, false},
		{output.Options{}, false},
	}
	for _, c := range cases {
		if got := c.opts.IsJSON(); got != c.want {
			t.Errorf("IsJSON(%+v) = %v; want %v", c.opts, got, c.want)
		}
	}
}
