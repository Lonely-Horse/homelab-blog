package markdown

import (
	"regexp"
	"strconv"
	"strings"
)

// 依旧是检测是否为fence类型
func isFenceLine(t string) bool {
	if strings.HasPrefix(t, "```") {
		return true
	}

	return false
}

// 检测是否为标题，以及他是几级标题
func isHeadingLine(t string) (int, bool) {
	var head string
	for i := 1; i <= 6; i++ {
		head = head + "#"
		if strings.HasPrefix(t, head+" ") {
			return i, true
		}
	}
	return 0, false
}

func headingText(t string, level int) string {
	text := t[level:]
	cand := strings.TrimRight(text, "#")
	if len(cand) != len(text) && cand[len(cand)-1] == ' ' {
		text = cand
	}
	return strings.TrimSpace(text)
}

// 检测是否为HR类型，也就是分隔线
var hrPattern = regexp.MustCompilePOSIX(`^[-]{3,}$`)

func isHRLine(t string) bool {
	return hrPattern.MatchString(t)
}

// 检测我 > 块
func isQuoteLine(t string) bool {
	return strings.HasPrefix(t, ">")
}

func stripQuotePrefix(t string) string {
	text := strings.TrimPrefix(t, ">")
	text = strings.TrimSpace(text)
	return text
}

// 检测fence的白名单
var fenceLangPattern = regexp.MustCompile(`^[a-zA-Z0-9+-]+$`)

func validateFenceLang(raw string) string {
	raw = strings.TrimSpace(raw)
	if !fenceLangPattern.MatchString(raw) {
		return ""
	}
	return raw
}

// 开始大量使用正则来完成字符的匹配，检测有序列表和无序列表
var orderedlistPattern = regexp.MustCompile(`^[0-9]+\. `)
var unorderedlistPattern = regexp.MustCompile(`^[\-*+] `)

func isListLine(t string) bool {
	if orderedlistPattern.MatchString(t) || unorderedlistPattern.MatchString(t) {
		return true
	}
	return false
}

func isOrderedList(t string) bool {
	return orderedlistPattern.MatchString(t)
}

// 切割出我们不需要的前缀，包括-，[x]，[ ]等等，只留下内容和我们扫描出来的具体类型
var tasklistPattern = regexp.MustCompile(`^\[[ x]\] \S`)

func stripListPrefix(t string) (string, listItemKind) {
	if !isListLine(t) {
		return t, listItemPlain
	}

	text := t[strings.IndexByte(t, ' ')+1:]

	if tasklistPattern.MatchString(text) {
		kind := listItemTaskOpen
		if text[1] == 'x' {
			kind = listItemTaskDone
		}
		text = text[4:]

		return text, kind
	}

	return text, listItemPlain
}

// 获取到列表初始数字
func istStartNum(t string) int {
	if !isOrderedList(t) {
		return 0
	}

	numstr := t[:strings.IndexByte(t, '.')]
	num, _ := strconv.Atoi(numstr)

	return num
}

// 完成我们的最终文本类型判定
func scanBlocks(lines []string) []block {
	var blocks []block
	inFence := false
	start := 0
	lang := ""
	paraStart := -1
	isQuote := false
	quoteStart := 0

	for i := 0; i < len(lines); i++ {
		t := strings.TrimSpace(lines[i])

		// > 守卫逻辑开始站岗，完成整个>块的完善
		if isQuote && !isQuoteLine(t) {
			b := block{kind: blockQuote}
			for _, line := range lines[quoteStart:i] {
				text := stripQuotePrefix(strings.TrimSpace(line))
				b.lines = append(b.lines, text)
			}
			blocks = append(blocks, b)
			isQuote = false
		}

		//判断是否为```
		if isFenceLine(t) {

			//先在判断是否存在```之前，完成正文内容的收尾
			if paraStart != -1 {
				b := block{kind: blockParagraph}
				b.lines = lines[paraStart:i]
				blocks = append(blocks, b)
				paraStart = -1
			}

			//判断这个是否为开栏```
			if !inFence {
				start = i
				lang = validateFenceLang(t[3:])
				inFence = true
				continue
			}

			//默认为闭栏```
			b := block{kind: blockFence, fence: lang}
			b.lines = lines[start+1 : i]
			blocks = append(blocks, b)
			inFence = false
			continue
		}

		if inFence {
			continue
		}

		if isQuoteLine(t) {
			if !isQuote {
				//先在判断是否存在>之前，完成正文内容的收尾
				if paraStart != -1 {
					b := block{kind: blockParagraph}
					b.lines = lines[paraStart:i]
					blocks = append(blocks, b)
					paraStart = -1
				}
				//定好>的起点
				quoteStart = i
				isQuote = true
			}
			continue
		}

		numb, ok := isHeadingLine(t)
		if ok {
			//在head之前判断一次是否存在正文未切割的情况
			if paraStart != -1 {
				b := block{kind: blockParagraph}
				b.lines = lines[paraStart:i]
				blocks = append(blocks, b)
				paraStart = -1
			}

			b := block{kind: blockHeading, level: numb}
			b.lines = append(b.lines, headingText(lines[i], numb))
			blocks = append(blocks, b)
			continue
		}

		//判断该行是否为HR，即---类
		if isHRLine(t) {
			//在head之前判断一次是否存在正文未切割的情况
			if paraStart != -1 {
				b := block{kind: blockParagraph}
				b.lines = lines[paraStart:i]
				blocks = append(blocks, b)
				paraStart = -1
			}

			//将我们的HR，也就是---部分直接添加到blocks中
			b := block{kind: blockHR}
			blocks = append(blocks, b)

			continue
		}

		//通过判定" "的位置，来切割段落
		if t == "" {
			if paraStart != -1 {
				b := block{kind: blockParagraph}
				b.lines = lines[paraStart:i]
				paraStart = -1
				blocks = append(blocks, b)
			}
			continue
		} else {
			//当我们的paraStart为-1的时候，也就是我们的段落起点位置，一旦下标开始往下移动的时候，这个-1条件就无法被满足，所以以此来达到定位段落的开头
			if paraStart == -1 {
				paraStart = i
			}
			continue
		}

	}
	if inFence {
		b := block{kind: blockFence, fence: lang}
		b.lines = lines[start+1:]
		blocks = append(blocks, b)
	}
	if paraStart != -1 {
		b := block{kind: blockParagraph}
		b.lines = lines[paraStart:]
		blocks = append(blocks, b)
	}
	if isQuote {
		b := block{kind: blockQuote}
		for _, line := range lines[quoteStart:] {
			text := stripQuotePrefix(strings.TrimSpace(line))
			b.lines = append(b.lines, text)
		}
		blocks = append(blocks, b)
		isQuote = false
	}

	return blocks
}
