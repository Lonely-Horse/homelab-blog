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

		case blockList:
			if b.ordered {
				fmt.Fprintf(&out, "<ol start=\"%d\">", b.listcode)
			} else {
				out.WriteString("<ul>")
			}

			for i, line := range b.lines {
				switch b.list[i] {
				case listItemPlain:
					out.WriteString("<li>" + renderInline(line) + "</li>")
				case listItemTaskOpen:
					out.WriteString("<li>" + `<input type="checkbox" disabled>` + renderInline(line) + "</li>")
				case listItemTaskDone:
					out.WriteString("<li>" + `<input type="checkbox" disabled checked>` + renderInline(line) + "</li>")
				}

			}
			if b.ordered {
				out.WriteString("</ol>\n")
			} else {
				out.WriteString("</ul>\n")
			}

		// 表格：契约规定 b.lines[0] 为表头行、b.lines[1:] 为数据行，
		// b.table 保存每列对齐方式，其长度 = 分隔行列数，也就是本表的"基准列数"
		case blockTable:
			out.WriteString("<table>\n<thead>\n")

			// 表头行：游标 i 走 b.table 而不是走 cells，这是 A 策略的落点——
			// 列数以分隔行为准，分隔行比表头多出的列渲染成空 <th>
			headerCells := splitTableRow(b.lines[0])
			out.WriteString("<tr>")
			for i := range b.table {
				// 取格前必须判越界：分隔行给了基准列数，但表头行是用户输入，格子数无保证
				cell := ""
				if i < len(headerCells) {
					cell = headerCells[i]
				}
				// 单元格内容走 renderInline，保证格内 `code`、**bold** 仍生效；
				// style 值来自 alignKind 枚举，属于内部可信数据，不需要转义
				fmt.Fprintf(&out, `<th style="text-align:%s">%s</th>`, b.table[i], renderInline(cell))
			}
			out.WriteString("</tr>\n</thead>\n<tbody>\n")

			// 数据行：必须逐行独立切分，因为每一行的 | 数量不一定与基准列数一致
			for _, row := range b.lines[1:] {
				rowCells := splitTableRow(row)
				out.WriteString("<tr>")
				for i := range b.table {
					// 同样按 A 策略：基准列数之外的多余格子直接丢弃，缺的补空 <td>
					cell := ""
					if i < len(rowCells) {
						cell = rowCells[i]
					}
					fmt.Fprintf(&out, `<td style="text-align:%s">%s</td>`, b.table[i], renderInline(cell))
				}
				out.WriteString("</tr>\n")
			}

			// 收尾的换行不能省，否则会和紧随其后的 <p>/<ul> 粘成一行
			out.WriteString("</tbody>\n</table>\n")

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
