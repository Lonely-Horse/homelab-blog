package markdown

import (
	"fmt"
)

type blockKind int
type alignKind int
type listItemKind int

const (
	blockParagraph blockKind = iota //段落无附加含义
	blockHeading                    //头部
	blockFence                      //语言表示，比如go，python等
	blockQuote                      //判定 > ，即引用块
	blockList                       //listItem类,即是列表
	blockTable                      //对齐方式
	blockHR                         //切割线类
)

const (
	alignLeft   alignKind = iota //左对齐
	alignCenter                  //居中
	alignRight                   //右对齐
)

const (
	listItemPlain    listItemKind = iota //作为头部为 - 的文本
	listItemTaskOpen                     //作为头部为 - [ ] 的文本
	listItemTaskDone                     //作为头部为 - [x] 的文本
)

// 这个block结构体是用于区分不同的类型的文本，比如说，列表list，正文lines等等，然后在外界调用的时候，根据block.kind能够合适的调用不同的方法和数据
type block struct {
	kind     blockKind
	lines    []string //扫描正文自动写入
	level    int
	fence    string
	list     []listItemKind //与lines进行一一对应，然后说明是否为普通项和任务项
	ordered  bool           //说明是否作为有序或者无序列表
	listcode int            //有序列表的起始编号，依靠ordered的判断
	table    []alignKind
}

func (k blockKind) String() string {
	switch k {
	case blockParagraph:
		return "paragraph"
	case blockHeading:
		return "heading"
	case blockFence:
		return "fence"
	case blockQuote:
		return "quote"
	case blockList:
		return "list"
	case blockTable:
		return "table"
	case blockHR:
		return "hr"
	default:
		return fmt.Sprintf("blockKind(%d)", int(k))
	}
}

// String 把对齐方式转成 CSS text-align 的取值，供渲染 <th>/<td> 的 style 使用。
// 返回值必须是 CSS 关键字 left/center/right，不能是 Go 的常量名：
// 写成 "alignLeft" 会渲染出 text-align:alignLeft，浏览器不认，三列全部退化成左对齐
func (a alignKind) String() string {
	switch a {
	case alignLeft:
		return "left"
	case alignCenter:
		return "center"
	case alignRight:
		return "right"
	default:
		return fmt.Sprintf("alignKind(%d)", int(a))
	}
}
