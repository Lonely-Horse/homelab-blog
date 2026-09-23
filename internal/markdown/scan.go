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
func listStartNum(t string) int {
	if !isOrderedList(t) {
		return 0
	}

	numstr := t[:strings.IndexByte(t, '.')]
	num, _ := strconv.Atoi(numstr)

	return num
}

// ===== 表格相关 =====
// 表格与本项目其他语法的最大区别：单看一行 | a | b | 无法判定它是表头还是普通正文，
// 必须"向前看一行"——下一行是分隔行才能确认。所以判定函数需要 lines 和下标，而不是单行文本。

// 分隔行中单个单元格的合法形态：--- / :--- / ---: / :---:
var tableDelimPattern = regexp.MustCompile(`^:?-{3,}:?$`)

// isTableRow 判定一行是否为表格行（表头行或数据行）：非空且含竖线。
// 含竖线这条约束是与水平线区分开的依据
func isTableRow(t string) bool {
	if t == "" {
		return false
	}

	return strings.Contains(t, "|")
}

// splitTableRow 按 | 切分一行，首尾竖线可省略，逐格去掉两端空白
func splitTableRow(t string) []string {
	t = strings.TrimSpace(t)
	t = strings.TrimPrefix(t, "|")
	t = strings.TrimSuffix(t, "|")

	parts := strings.Split(t, "|")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}

	return parts
}

// isTableDelim 判定分隔行：必须含竖线，且每一格都是 tableDelimPattern 的形态。
// 含竖线是硬要求：否则不带竖线的 --- 会被 isHRLine 抢先匹配成水平线，表格永远认不出来
func isTableDelim(t string) bool {
	if !strings.Contains(t, "|") {
		return false
	}

	cells := splitTableRow(t)
	if len(cells) == 0 {
		return false
	}

	for _, c := range cells {
		if !tableDelimPattern.MatchString(c) {
			return false
		}
	}

	return true
}

// isTableStart 表格起点判定，这是全项目唯一需要"向前看一行"的判定函数：
// lines[i] 必须是表格行，lines[i+1] 必须是分隔行，两者同时成立才算表格起点。
// 只看 lines[i] 会把普通正文里含 | 的一行误判成表头
func isTableStart(lines []string, i int) bool {
	// 向前看一行之前先防越界：末行不可能是表头，因为它没有下一行来确认
	if i+1 >= len(lines) {
		return false
	}

	if !isTableRow(strings.TrimSpace(lines[i])) {
		return false
	}

	return isTableDelim(strings.TrimSpace(lines[i+1]))
}

// parseAlign 把分隔行中的单个单元格解析为对齐方式：
// :--- 左对齐、:---: 居中、---: 右对齐，无冒号按左对齐处理（契约已定）
func parseAlign(cell string) alignKind {
	left := strings.HasPrefix(cell, ":")
	right := strings.HasSuffix(cell, ":")

	switch {
	case left && right:
		return alignCenter
	case right:
		return alignRight
	default:
		return alignLeft
	}
}

// parseTableAlign 把整个分隔行解析成每列的对齐方式，返回长度即基准列数
func parseTableAlign(delim string) []alignKind {
	cells := splitTableRow(delim)
	aligns := make([]alignKind, 0, len(cells))

	for _, c := range cells {
		aligns = append(aligns, parseAlign(c))
	}

	return aligns
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

	isList := false
	listOrdered := false
	listNum := 0
	listStart := 0

	isTable := false
	tableStart := 0

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

		// 列表守卫逻辑
		// 当isList为true，也就是说这是已经有首部，而缺少尾部的列表
		// !isListLine(t)，该行并非是列表
		// isOrderedList(t)) != listOrdered，存在有序列表首部，缺失尾部
		if isList && (!isListLine(t) || isOrderedList(t) != listOrdered) {
			b := block{kind: blockList, ordered: listOrdered, listcode: listNum}

			for _, line := range lines[listStart:i] {
				text, kind := stripListPrefix(strings.TrimSpace(line))
				b.lines = append(b.lines, text)
				b.list = append(b.list, kind)
			}

			blocks = append(blocks, b)
			isList = false
		}

		// 表格守卫逻辑
		// 与列表不同，表格没有"类型变化"需要比较，只有进表和出表两种状态：
		// 当前行不再是表格行（不含竖线或为空行），就把已攒下的表格收尾
		if isTable && !isTableRow(t) {
			b := block{kind: blockTable}

			// 第 0 行固定是表头行
			b.lines = append(b.lines, lines[tableStart])

			// 从 tableStart+2 开始才是数据行，+2 是为了跳过表头和分隔行本身
			for _, line := range lines[tableStart+2 : i] {
				b.lines = append(b.lines, line)
			}

			// 对齐方式从分隔行解析，其长度就是本表的基准列数
			b.table = parseTableAlign(lines[tableStart+1])

			blocks = append(blocks, b)
			isTable = false
		}

		// 已经进表：本行内容留到收尾时统一收集，这里不做任何其他规则判定。
		// 尤其不能让它落到下面的段落分支，否则分隔行会被当成段落起点，
		// 表头被拆成两段，整块降级成"奇怪的半表格"
		if isTable {
			continue
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

		//现在是第一次捕获到list，开始记录列表长度和类型
		if isListLine(t) {
			if paraStart != -1 {
				b := block{kind: blockParagraph}
				b.lines = lines[paraStart:i]
				blocks = append(blocks, b)
				paraStart = -1
			}

			if !isList {
				listStart = i
				listOrdered = isOrderedList(t)
				listNum = listStartNum(t)
				isList = true
			}

			continue
		}

		// 表格起点判定：必须"向前看一行"，lines[i] 是表头、lines[i+1] 是分隔行
		// 放在段落分支之前：否则表头行会被段落分支先接走，表格永远建不起来
		if isTableStart(lines, i) {
			if paraStart != -1 {
				b := block{kind: blockParagraph}
				b.lines = lines[paraStart:i]
				blocks = append(blocks, b)
				paraStart = -1
			}

			// 只记录起点，行内容等收尾时统一从 lines 里取（与列表同一套路）
			tableStart = i
			isTable = true
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

	if isTable {
		b := block{kind: blockTable}
		b.lines = append(b.lines, lines[tableStart])

		// 循环走到末尾仍未出表：数据行一直取到最后一行，同样跳过 tableStart+1 的分隔行
		for _, line := range lines[tableStart+2:] {
			b.lines = append(b.lines, line)
		}

		b.table = parseTableAlign(lines[tableStart+1])
		blocks = append(blocks, b)
		isTable = false
	}

	if paraStart != -1 {
		b := block{kind: blockParagraph}
		b.lines = lines[paraStart:]
		blocks = append(blocks, b)
	}

	if isList {
		b := block{kind: blockList, ordered: listOrdered, listcode: listNum}
		for _, line := range lines[listStart:] {
			text, kind := stripListPrefix(strings.TrimSpace(line))
			b.lines = append(b.lines, text)
			b.list = append(b.list, kind)
		}
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
