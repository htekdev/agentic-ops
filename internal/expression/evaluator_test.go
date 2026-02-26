package expression

import (
	"testing"
)

func TestContextEvaluate(t *testing.T) {
	ctx := NewContext()
	ctx.Event["cwd"] = "/test/path"
	ctx.Event["file"] = map[string]interface{}{
		"path":   "src/main.go",
		"action": "edit",
	}
	ctx.Env["NODE_ENV"] = "development"

	tests := []struct {
		name    string
		expr    string
		want    interface{}
		wantErr bool
	}{
		{
			name: "literal true",
			expr: "true",
			want: true,
		},
		{
			name: "literal false",
			expr: "false",
			want: false,
		},
		{
			name: "literal null",
			expr: "null",
			want: nil,
		},
		{
			name: "literal number",
			expr: "42",
			want: int64(42),
		},
		{
			name: "literal string",
			expr: "'hello'",
			want: "hello",
		},
		{
			name: "event property",
			expr: "event.cwd",
			want: "/test/path",
		},
		{
			name: "nested property",
			expr: "event.file.path",
			want: "src/main.go",
		},
		{
			name: "env access",
			expr: "env.NODE_ENV",
			want: "development",
		},
		{
			name: "equality true",
			expr: "'a' == 'a'",
			want: true,
		},
		{
			name: "equality false",
			expr: "'a' == 'b'",
			want: false,
		},
		{
			name: "inequality",
			expr: "'a' != 'b'",
			want: true,
		},
		{
			name: "case insensitive equality",
			expr: "'Hello' == 'hello'",
			want: true,
		},
		{
			name: "logical and",
			expr: "true && true",
			want: true,
		},
		{
			name: "logical or",
			expr: "false || true",
			want: true,
		},
		{
			name: "logical not",
			expr: "!false",
			want: true,
		},
		{
			name: "comparison less than",
			expr: "1 < 2",
			want: true,
		},
		{
			name: "comparison greater than",
			expr: "2 > 1",
			want: true,
		},
		{
			name: "parentheses",
			expr: "(1 < 2) && (3 > 2)",
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ctx.Evaluate(tt.expr)
			if (err != nil) != tt.wantErr {
				t.Errorf("Evaluate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Evaluate() = %v (%T), want %v (%T)", got, got, tt.want, tt.want)
			}
		})
	}
}

func TestContextEvaluateString(t *testing.T) {
	ctx := NewContext()
	ctx.Event["file"] = map[string]interface{}{
		"path": "test.js",
	}

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "no expressions",
			input: "just text",
			want:  "just text",
		},
		{
			name:  "single expression",
			input: "path: ${{ event.file.path }}",
			want:  "path: test.js",
		},
		{
			name:  "multiple expressions",
			input: "${{ event.file.path }} is a ${{ 'file' }}",
			want:  "test.js is a file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ctx.EvaluateString(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("EvaluateString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("EvaluateString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestContextEvaluateBool(t *testing.T) {
	ctx := NewContext()

	tests := []struct {
		name    string
		expr    string
		want    bool
		wantErr bool
	}{
		{"true", "true", true, false},
		{"false", "false", false, false},
		{"comparison", "1 == 1", true, false},
		{"and", "true && false", false, false},
		{"or", "true || false", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ctx.EvaluateBool(tt.expr)
			if (err != nil) != tt.wantErr {
				t.Errorf("EvaluateBool() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("EvaluateBool() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuiltinContains(t *testing.T) {
	tests := []struct {
		name    string
		search  interface{}
		item    string
		want    bool
		wantErr bool
	}{
		{
			name:   "string contains",
			search: "Hello World",
			item:   "World",
			want:   true,
		},
		{
			name:   "string not contains",
			search: "Hello World",
			item:   "Foo",
			want:   false,
		},
		{
			name:   "case insensitive",
			search: "Hello World",
			item:   "world",
			want:   true,
		},
		{
			name:   "array contains",
			search: []interface{}{"a", "b", "c"},
			item:   "b",
			want:   true,
		},
		{
			name:   "array not contains",
			search: []interface{}{"a", "b", "c"},
			item:   "d",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := builtinContains(tt.search, tt.item)
			if (err != nil) != tt.wantErr {
				t.Errorf("builtinContains() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("builtinContains() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuiltinStartsWith(t *testing.T) {
	tests := []struct {
		str    string
		prefix string
		want   bool
	}{
		{"Hello World", "Hello", true},
		{"Hello World", "World", false},
		{"Hello World", "hello", true}, // case insensitive
	}

	for _, tt := range tests {
		t.Run(tt.str+"_"+tt.prefix, func(t *testing.T) {
			got, err := builtinStartsWith(tt.str, tt.prefix)
			if err != nil {
				t.Errorf("builtinStartsWith() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("builtinStartsWith() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuiltinEndsWith(t *testing.T) {
	tests := []struct {
		str    string
		suffix string
		want   bool
	}{
		{"Hello World", "World", true},
		{"Hello World", "Hello", false},
		{"test.js", ".js", true},
	}

	for _, tt := range tests {
		t.Run(tt.str+"_"+tt.suffix, func(t *testing.T) {
			got, err := builtinEndsWith(tt.str, tt.suffix)
			if err != nil {
				t.Errorf("builtinEndsWith() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("builtinEndsWith() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuiltinFormat(t *testing.T) {
	tests := []struct {
		name string
		args []interface{}
		want string
	}{
		{
			name: "simple format",
			args: []interface{}{"Hello {0}", "World"},
			want: "Hello World",
		},
		{
			name: "multiple placeholders",
			args: []interface{}{"{0} {1} {2}", "a", "b", "c"},
			want: "a b c",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := builtinFormat(tt.args...)
			if err != nil {
				t.Errorf("builtinFormat() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("builtinFormat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuiltinJoin(t *testing.T) {
	tests := []struct {
		name string
		arr  []interface{}
		sep  string
		want string
	}{
		{
			name: "default separator",
			arr:  []interface{}{"a", "b", "c"},
			sep:  "",
			want: "a,b,c",
		},
		{
			name: "custom separator",
			arr:  []interface{}{"a", "b", "c"},
			sep:  " - ",
			want: "a - b - c",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got interface{}
			var err error
			if tt.sep == "" {
				got, err = builtinJoin(tt.arr)
			} else {
				got, err = builtinJoin(tt.arr, tt.sep)
			}
			if err != nil {
				t.Errorf("builtinJoin() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("builtinJoin() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuiltinToJSON(t *testing.T) {
	got, err := builtinToJSON(map[string]interface{}{"key": "value"})
	if err != nil {
		t.Errorf("builtinToJSON() error = %v", err)
		return
	}
	// JSON may have different spacing, just check it's valid
	if got != `{"key":"value"}` {
		t.Errorf("builtinToJSON() = %v", got)
	}
}

func TestBuiltinFromJSON(t *testing.T) {
	got, err := builtinFromJSON(`{"key":"value"}`)
	if err != nil {
		t.Errorf("builtinFromJSON() error = %v", err)
		return
	}
	m, ok := got.(map[string]interface{})
	if !ok {
		t.Errorf("builtinFromJSON() returned %T, want map", got)
		return
	}
	if m["key"] != "value" {
		t.Errorf("builtinFromJSON() key = %v, want 'value'", m["key"])
	}
}

func TestFunctionCallInContext(t *testing.T) {
	ctx := NewContext()
	ctx.Event["file"] = map[string]interface{}{
		"path": "src/utils/helper.js",
	}

	tests := []struct {
		name    string
		expr    string
		want    interface{}
		wantErr bool
	}{
		{
			name: "contains function",
			expr: "contains(event.file.path, 'utils')",
			want: true,
		},
		{
			name: "startsWith function",
			expr: "startsWith(event.file.path, 'src')",
			want: true,
		},
		{
			name: "endsWith function",
			expr: "endsWith(event.file.path, '.js')",
			want: true,
		},
		{
			name: "always function",
			expr: "always()",
			want: true,
		},
		{
			name: "nested function in condition",
			expr: "contains(event.file.path, 'utils') && endsWith(event.file.path, '.js')",
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ctx.Evaluate(tt.expr)
			if (err != nil) != tt.wantErr {
				t.Errorf("Evaluate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Evaluate() = %v, want %v", got, tt.want)
			}
		})
	}
}
