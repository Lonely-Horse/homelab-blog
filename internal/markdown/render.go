package markdown

import (
	"fmt"
	"html"
	"log"
	"strings"
)

func renderBlocks(bs []block) string {
	var out strings.Builder
	for _, b := range bs {
		switch b.kind {

		//当文本为段落的时候，触发
		case blockParagraph:
			out.WriteString("<p>")
			for i, line := range b.lines {
				if i > 0 {
					out.WriteString("\n")
				}
				out.WriteString(renderInline(line))
			}
			out.WriteString("</p>\n")

		//当文本为代码块，也就是fence时触发
		case blockFence:
			if b.fence == "" {
				out.WriteString(`<pre><code class="language-text`)
				out.WriteString("\">")
			} else {
				out.WriteString(`<pre><code class="language-`)
				out.WriteString(html.EscapeString(b.fence))
				out.WriteString("\">")
			}
			for i, line := range b.lines {
				if i > 0 {
					out.WriteString("\n")
				}
				out.WriteString(html.EscapeString(line))
			}
			out.WriteString("</code></pre>\n")

		//当文本为多级标题的时候，触发
		case blockHeading:
			fmt.Fprintf(&out, "<h%d>%s</h%d>\n", b.level, html.EscapeString(b.lines[0]), b.level)

		//该行属于---的时候，我们完成转义
		case blockHR:
			out.WriteString("<hr>\n")

		//该行为>类，触发
		case blockQuote:
			out.WriteString("<blockquote>")
			for i, line := range b.lines {
				if i > 0 {
					out.WriteString("\n")
				}
				out.WriteString(renderInline(line))
			}
			out.WriteString("</blockquote>\n")
		default:
			out.WriteString("<p>")
			for i, line := range b.lines {
				if i > 0 {
					out.WriteString("\n")
				}
				out.WriteString(html.EscapeString(line))
			}
			out.WriteString("</p>\n")
			log.Printf("The %v kind didn't find", b.kind)
		}
	}
	return out.String()
}
