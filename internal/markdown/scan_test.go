package markdown

import (
	"reflect"
	"testing"
)

func TestScanBlocks(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  []block
	}{
		{
			name:  "单个普通段落",
			input: []string{"The test"},
			want:  []block{{kind: blockParagraph, lines: []string{"The test"}}},
		},
		{
			name:  "段落后紧接代码块",
			input: []string{"The test", "```go", "func main() {", "```"},
			want: []block{
				{kind: blockParagraph, lines: []string{"The test"}},
				{kind: blockFence, fence: "go", lines: []string{"func main() {"}},
			},
		},
		{
			name:  "未闭合围栏",
			input: []string{"```go", "func main() {", "fmt.Printf('The test')", "}"},
			want:  []block{{kind: blockFence, fence: "go", lines: []string{"func main() {", "fmt.Printf('The test')", "}"}}},
		},
		{
			name:  "空行切出两个段落",
			input: []string{"第一段", "", "", "第二段"},
			want: []block{
				{kind: blockParagraph, lines: []string{"第一段"}},
				{kind: blockParagraph, lines: []string{"第二段"}},
			},
		},
		{
			name:  "围栏内的空行不算段落分隔",
			input: []string{"```go", "a", "", "b", "```"},
			want:  []block{{kind: blockFence, fence: "go", lines: []string{"a", "", "b"}}},
		},
		{
			name:  "空输入",
			input: []string{},
			want:  nil,
		},
		{
			name:  "只有空行",
			input: []string{"", "   ", ""},
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := scanBlocks(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
		})
	}

}
