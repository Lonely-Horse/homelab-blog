package markdown

import (
	"html"
	"strings"
)

func sanitizeURL(url string) bool {
	whitelist := make([]string, 0)
	whitelist = append(whitelist, "http://", "https://")

	for _, prefix := range whitelist {
		if strings.HasPrefix(url, prefix) {
			return true
		}
	}

	return false
}

func renderInline(line string) string {
	var out strings.Builder

	switch {
	case strings.Contains(line, "`") && strings.Count(line, "`")%2 == 0:
		texts := strings.Split(line, "`")
		for i, text := range texts {
			if i%2 != 0 {
				out.WriteString("<code>" + html.EscapeString(text) + "</code>")
			} else {
				out.WriteString(renderInline(text))
			}
		}
		return out.String()
	case strings.Contains(line, "**") && strings.Count(line, "**")%2 == 0:
		texts := strings.Split(line, "**")
		for i, text := range texts {
			if i%2 != 0 {
				out.WriteString("<strong>" + html.EscapeString(text) + "</strong>")
			} else {
				out.WriteString(renderInline(text))
			}
		}
		return out.String()
	case strings.Contains(line, "["):
		openA := strings.Index(line, "[")
		closeA := strings.Index(line[openA+1:], "]")
		lineA := openA + 1 + closeA

		if closeA == -1 {
			out.WriteString(html.EscapeString(line))
			return out.String()
		}
		if !strings.HasPrefix(line[lineA+1:], "(") {
			out.WriteString(html.EscapeString(line))
			return out.String()
		}

		openB := strings.Index(line, "(")
		closeB := strings.Index(line[openB+1:], ")")
		lineB := openB + 1 + closeB

		if closeB == -1 {
			out.WriteString(html.EscapeString(line))
			return out.String()
		}
		text := line[openA+1 : lineA]
		url := line[openB+1 : lineB]

		if !sanitizeURL(url) {
			out.WriteString(html.EscapeString(line))
			return out.String()
		}

		out.WriteString(renderInline(line[:openA]))
		out.WriteString(`<a href="` + html.EscapeString(url) + `">` + html.EscapeString(text) + "</a>")
		out.WriteString(renderInline(line[lineB+1:]))

		return out.String()
	default:
		out.WriteString(html.EscapeString(line))
		return out.String()
	}
}
