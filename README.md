# LonelyHorse 个人博客

用 Go 标准库从零手写的个人技术博客。单作者、单管理员，目标不是功能多，而是**每一行代码都能讲清楚**。

- 学习方向：Go · 云原生 · DevOps
- 技术取向：标准库优先，全项目只有一个第三方依赖
- 部署取向：单机 + SQLite，不引入重组件（不要 MySQL / Redis / 消息队列）

## 技术栈

| 维度 | 选型 |
| :--- | :--- |
| 语言 | Go 1.25（需 ≥ 1.24，用到标准库 `crypto/pbkdf2`） |
| HTTP | `net/http` + Go 1.22+ `ServeMux`（原生方法 + 路径变量路由） |
| 模板 | `html/template`（服务端渲染 SSR） |
| 数据库 | SQLite（WAL 模式） |
| 驱动 | `github.com/glebarez/go-sqlite`（纯 Go，无需 CGO）——**全项目唯一的第三方依赖** |
| Markdown | 自研解析器（纯标准库） |
| 密码 | PBKDF2-HMAC-SHA256（标准库），会话存 SQLite |

## 目录结构（简）

```text
homelab-blog/
├── main.go                 # 仅调用 app.Run()
├── cmd/admin-seed/         # 一次性创建/重置管理员账号
├── internal/
│   ├── app/                # 组合根：装配依赖并启动
│   ├── config/             # 全部配置默认值
│   ├── db/                 # SQLite 连接、建表、幂等迁移（schema.sql 内嵌）
│   ├── posts/              # 文章领域（SQL 出口）
│   ├── profile/            # 关于页（站主简介）
│   ├── auth/               # 密码派生 + 会话
│   ├── markdown/           # 自研解析器（纯函数，零内部依赖）
│   └── server/             # 传输层：路由、中间件、Handler
├── web/
│   ├── templates/          # 页面模板（编译期内嵌进二进制）
│   └── static/             # CSS / JS（编译期内嵌进二进制）
└── data/                   # 运行时数据（sqlite 库），不进版本库
```

## 快速开始

前置：Go ≥ 1.24。

**1. 启动服务**

```bash
go run .
```

默认监听 `127.0.0.1:5200`，数据库自动生成在 `./data/blog.db`（首次启动自动建表并自检 PRAGMA）。

**2. 创建管理员（首次必须，否则无法登录）**

```bash
go run ./cmd/admin-seed -user admin
# 随后按提示从标准输入输入密码
```

账号已存在时默认拒绝覆盖；确需改密码追加 `-reset`（会同时吊销该账号的所有会话）。

**3. 访问**

| 入口 | 地址 |
| :--- | :--- |
| 博客列表 | http://127.0.0.1:5200/posts |
| 幕布首页 | http://127.0.0.1:5200/ |
| 后台登录 | http://127.0.0.1:5200/admin/login |

## 配置

目前**只有一个环境变量**：

| 变量 | 默认 | 说明 |
| :--- | :--- | :--- |
| `SESSION_COOKIE_SECURE` | `false` | 用 HTTPS 访问时**必须设为 `true`**，否则会话 Cookie 在应用层是明文 |

其余配置（端口、各类超时、限流阈值、PBKDF2 迭代次数等）都在 [internal/config/config.go](internal/config/config.go) 中以默认值写死。

## 路由一览

**公开（无需登录）**

| 方法 | 路径 | 说明 |
| :--- | :--- | :--- |
| GET | `/` | 幕布首页（一次进入动效） |
| GET | `/posts` | 博客列表（仅已发布，按发布时间倒序） |
| GET | `/post/{slug}` | 文章详情（草稿返回 404） |
| GET | `/about` | 关于页 |
| GET | `/healthz` | 存活探针（不查库、不鉴权） |
| GET | `/static/` | 静态资源 |

**管理（需登录）**

| 方法 | 路径 | 说明 |
| :--- | :--- | :--- |
| GET / POST | `/admin/login` | 登录页 / 提交登录 |
| POST | `/admin/logout` | 退出（无需登录也能清 Cookie） |
| GET | `/admin` | 后台首页 |
| GET | `/admin/posts` | 文章列表（含草稿） |
| GET | `/admin/posts/new` | 新建文章 |
| POST | `/admin/posts` | 创建文章 |
| GET | `/admin/posts/{id}/edit` | 编辑文章 |
| POST | `/admin/posts/{id}` | 更新文章 |
| DELETE | `/admin/posts/{id}` | 删除文章（前端用 fetch 调用） |
| GET / POST | `/about/edit` | 编辑站主简介 |

## 已实现 / 规划中

**已实现**

- 基础设施：配置加载、SQLite 自动建表、PRAGMA 自检、`user_version` 幂等迁移
- 自研 Markdown 解析器：标题 / 围栏代码块 / 引用 / 列表 / 任务列表 / 表格 / 水平线 / 粗体 / 斜体 / 行内代码 / 链接 / 图片
- 认证：PBKDF2 加盐、会话表、页面 / API 双中间件、安全响应头（CSP 等）、登录限流（按客户端 IP 的内存令牌桶）
- 文章：草稿 / 发布、列表、详情、新建 / 编辑 / 删除
- 关于页：站主简介在线编辑
- 前端：SSR 模板、明暗主题切换、右上角常驻浮动返回按钮

**规划中 / 未做**

- 项目库 `/projects`（导航已有入口，点击目前 404）
- 图片上传 `/admin/upload` 与 Markdown 实时预览 `/admin/preview`
- 代码高亮（代码块目前无语法着色）
- 单元测试（解析器等尚无 `_test.go`）
- Docker 多阶段镜像（尚未编写 `Dockerfile`）
- `cmd/render-backfill`（解析器升级后的历史 HTML 回填命令）

## 安全设计要点

- **后台不裸露公网**：生产环境在 Nginx 用 `location ^~ /admin` 只放行 Tailscale 网段，Go 层双中间件作第二道防线
- **双中间件**：页面路由未登录 → 303 跳登录页；fetch 数据路由未登录 → 401 JSON（避免 fetch 跟随重定向被误判为成功）
- **XSS 防护**：CSP 禁内联；Markdown 输出的所有文本强制 HTML 转义；链接 / 图片 URL 走 scheme 白名单
- **Cookie**：`HttpOnly` + `SameSite=Lax`；HTTPS 下 `Secure=true`
- **登录限流**：防 PBKDF2 被当作 CPU 放大点打满，同时抑制暴力猜密

## 相关文档

- [技术文档.md](技术文档.md)——完整设计：系统架构、路由表、数据模型、解析器铁律、部署方案、验收清单、决策记录（ADR）
