package markdown

import "testing"

func TestRender(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "两个空输入",
			in:   "第一段\n\n第二段",
			want: "<p>第一段</p>\n<p>第二段</p>\n",
		},
		{
			name: "多行属于同一段",
			in:   "hello\nworld",
			want: "<p>hello\nworld</p>\n",
		},
		{
			name: "文字里的特殊符号会被转义",
			in:   "\n\na < b & c\n\n",
			want: "<p>a &lt; b &amp; c</p>\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Render(tt.in)
			if got != tt.want {
				t.Errorf("Render(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
