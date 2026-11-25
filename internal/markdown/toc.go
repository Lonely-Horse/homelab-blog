package markdown

import (
	"fmt"
	"html"
	"strconv"
	"strings"
	"unicode"
)

type Heading struct {
	Level int
	Text  string
	ID    string
}

type renderer struct {
	headings []Heading
	seen     map[string]int
	seq      int
}

func slugify(s string) string {
	var b strings.Builder
	pendingSep := false

	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if pendingSep && b.Len() != 0 {
				b.WriteString("-")
			}
			pendingSep = false
			b.WriteRune(unicode.ToLower(r))
			continue
		}
		pendingSep = true
	}

	return b.String()
}

func (r *renderer) heading(b block) string {
	text := b.lines[0]

	if b.level < 2 || b.level > 4 {
		return fmt.Sprintf("<h%d>%s</h%d>\n", b.level, html.EscapeString(text), b.level)
	}

	base := slugify(text)

	if base == "" {
		r.seq++
		base = "sec-" + strconv.Itoa(r.seq)
	}

	count := r.seen[base]
	r.seen[base] = count + 1

	id := base
	if count > 0 {
		id = base + "-" + strconv.Itoa(count+1)
	}

	r.headings = append(r.headings, Heading{Level: b.level, Text: text, ID: id})

	return fmt.Sprintf("<h%d id=\"%s\">%s</h%d>\n", b.level, id, html.EscapeString(text), b.level)
}

type tocNode struct {
	Heading
	Children []*tocNode
}

func writeTOCList(b *strings.Builder, nodes []*tocNode) {
	b.WriteString("<ol>")
	for _, n := range nodes {
		b.WriteString("<li>")
		b.WriteString(`<a href="#` + n.ID + `">` + html.EscapeString(n.Text) + "</a>")
		if len(n.Children) > 0 {
			writeTOCList(b, n.Children)
		}
		b.WriteString("</li>")
	}
	b.WriteString("</ol>")
}

func RenderTOC(hs []Heading) string {
	if len(hs) < 3 {
		return ""
	}

	var roots []*tocNode
	var stack []*tocNode

	for i := range hs {
		n := &tocNode{Heading: hs[i]}

		for len(stack) > 0 && stack[len(stack)-1].Level >= n.Level {
			stack = stack[:len(stack)-1]
		}

		if len(stack) == 0 {
			roots = append(roots, n)
		} else {
			stack[len(stack)-1].Children = append(stack[len(stack)-1].Children, n)
		}
		stack = append(stack, n)
	}

	var b strings.Builder

	writeTOCList(&b, roots)

	return `<nav class="toc-nav">` + b.String() + `</nav>`
}
