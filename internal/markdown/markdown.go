package markdown

import (
	"strings"
)

func preprocess(src string) string {
	return strings.ReplaceAll(src, "\r\n", "\n")
}

func splitParagraphs(src string) []string {
	var blocks []string
	var current []string

	for _, line := range strings.Split(src, "\n") {
		if strings.TrimSpace(line) == "" {
			if len(current) > 0 {
				blocks = append(blocks, strings.Join(current, "\n"))
				current = nil
			}
			continue
		}

		current = append(current, line)

	}
	if len(current) > 0 {
		blocks = append(blocks, strings.Join(current, "\n"))
	}
	return blocks
}

func Render(src string) string {
	src = preprocess(src)
	lines := strings.Split(src, "\n")
	return renderBlocks(scanBlocks(lines))

}
