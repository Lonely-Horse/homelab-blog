//go:build ignore

// 临时验证脚本：插入一篇带多级标题的演示文章，用于前端 TOC 视觉验收。
// 用毕即删（go run tmp_toc_demo.go），随后用 tmp_toc_demo_clean.go 清理。
package main

import (
	"fmt"
	"homelab-blog/internal/config"
	"homelab-blog/internal/db"
	"homelab-blog/internal/markdown"
	"homelab-blog/internal/posts"
	"log"
)

const demoMD = `## 下载

### 下载地址

正文段落，验证目录与锚点。

### 版本的选择

更多正文内容，撑开滚动距离。

## 安装

### Windows 安装

内容。

### Linux 下安装

内容。

#### 检查环境变量

深层四级标题。

### Mac 下安装

内容。

## GOROOT 和 GOPATH

内容。

## Go 开发编辑器

### VS Code 介绍

内容。

#### 安装中文简体插件

内容。

#### 安装 go 扩展

内容。

## 第一个 Go 程序

### Hello World

内容。

### go mod init

内容。
`

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	database, err := db.Open(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	store := posts.NewStore(cfg, database)
	body, headings := markdown.Render(demoMD)
	id, err := store.Create(&posts.Post{
		Title:       "TOC 演示文章（临时）",
		Slug:        "toc-demo",
		Summary:     "用于前端目录视觉验收，验完即删。",
		ContentMD:   demoMD,
		ContentHTML: body,
		TOCHTML:     markdown.RenderTOC(headings),
		Status:      "published",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("created post id =", id)
}
