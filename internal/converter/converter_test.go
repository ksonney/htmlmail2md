package converter

import "testing"

func TestConvert(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "heading and paragraph",
			input: `<html><body><h1>Hello</h1><p>World</p></body></html>`,
			want:  "# Hello\n\nWorld\n",
		},
		{
			name:  "inline formatting",
			input: `<html><body><p>Hello <b>World</b>!</p></body></html>`,
			want:  "Hello **World**!\n",
		},
		{
			name:  "link",
			input: `<html><body><p>See <a href="https://example.com">the docs</a></p></body></html>`,
			want:  "See [the docs](https://example.com)\n",
		},
		{
			name:  "unordered list",
			input: `<html><body><ul><li>one</li><li>two</li></ul></body></html>`,
			want:  "- one\n- two\n",
		},
		{
			name:  "nested unordered list",
			input: `<html><body><ul><li>one</li><li><ul><li>two</li></ul></li></ul></body></html>`,
			want:  "- one\n  - two\n",
		},
		{
			name:  "blockquote",
			input: `<html><body><blockquote><p>Quote</p></blockquote></body></html>`,
			want:  "> Quote\n",
		},
		{
			name:  "code block",
			input: "<html><body><pre><code>fmt.Println(\"hi\")</code></pre></body></html>",
			want:  "```\nfmt.Println(\"hi\")\n```\n",
		},
		{
			name:  "skips head and style",
			input: `<html><head><style>body { color: red }</style><title>Hi</title></head><body><p>Hi</p></body></html>`,
			want:  "Hi\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Convert([]byte(tt.input))
			if err != nil {
				t.Fatalf("Convert() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("Convert() =\n%q\nwant\n%q", got, tt.want)
			}
		})
	}
}
