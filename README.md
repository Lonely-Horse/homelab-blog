# homelab-blog

用 Go 标准库从零手写的个人技术博客。单作者、单管理员，目标不是"功能多"，而是**每一行代码都能讲清楚**。

- 学习方向：Go · 云原生 · DevOps
- 技术取向：标准库优先，全项目只有一个第三方依赖
- 部署取向：单机 + SQLite，不引入重组件（不要 MySQL / Redis / 消息队列）

## 功能现状

| 模块 | 状态 | 说明 |
| :--- | :--- | :--- |
| 文章 | 已实现 | 草稿 / 发布、列表、详情、新建、编辑、删除 |
| 项目库 | 已实现 | 项目卡片、标签、封面、状态、排序、置顶、附件（ZIP）下载 |
| 关于页 | 已实现 | 站主简介在线编辑 |
| 认证 | 已实现 | PBKDF2 加盐 + 会话表 + 页面 / API 双中间件 |
| Markdown | 已实现 | 自研解析器（纯标准库），渲染期产 HTML 并入库 |
| 静态资源与模板 | 已实现 | 编译期 `//go:embed` 内嵌进二进制，无需外部文件 |
| 离线部署 | 已实现 | 多阶段 Docker 镜像（非 root 运行）+ docker-compose |
| 安全响应头 | 已实现 | CSP / X-Frame-Options / nosniff / Referrer-Policy |
| 登录限流 | 已实现 | 按客户端 IP 的内存令牌桶 |

## 技术栈

| 维度 | 选型 |
| :--- | :--- |
| 语言 | Go 1.25（`go.mod` 声明 1.25.0，使用标准库 `crypto/pbkdf2`） |
| HTTP | `net/http` + Go 1.22+ `ServeMux`（原生方法 + 路径变量路由） |
| 模板 | `html/template`（服务端渲染 SSR） |
| 数据库 | SQLite（WAL 模式，`foreign_keys=ON`） |
| 驱动 | `github.com/glebarez/go-sqlite`（纯 Go，无需 CGO）——**全项目唯一的第三方依赖** |
| Markdown | 自研解析器（纯标准库） |
| 密码 | PBKDF2-HMAC-SHA256（标准库），会话落 SQLite |

## 目录结构

```text
homelab-blog/
├── main.go                 # 仅调用 app.Run()
├── cmd/admin-seed/         # 一次性创建 / 重置管理员账号
├── internal/
│   ├── app/                # 组合根：装配依赖并启动
│   ├── config/             # 全部配置默认值
│   ├── db/                 # SQLite 连接、建表、幂等迁移（schema.sql 内嵌）
│   ├── posts/              # 文章领域（SQL 出口）
│   ├── projects/           # 项目库领域
│   ├── profile/            # 关于页（站主简介）
│   ├── auth/               # 密码派生 + 会话
│   ├── markdown/           # 自研解析器（纯函数，零内部依赖）
│   └── server/             # 传输层：路由、中间件、Handler
├── web/
│   ├── templates/          # 页面模板（编译期内嵌进二进制）
│   ├── static/             # CSS / JS（编译期内嵌进二进制）
│   └── embed.go            # //go:embed templates static
├── docker/
│   ├── Dockerfile          # 多阶段构建，非 root 运行
│   └── docker-compose.yaml
└── data/                   # 运行时数据（sqlite 库 + 上传附件），不进版本库
```

## 快速开始

前置：Go ≥ 1.25。

**1. 启动服务**

```bash
go run .
```

默认监听 `0.0.0.0:5200`，数据库自动生成在 `./data/blog.db`（首次启动自动建表并自检 PRAGMA）。

**2. 创建管理员（首次必须，否则无法登录）**

```bash
go run ./cmd/admin-seed -user admin
# 随后按提示从标准输入输入密码
```

账号已存在时默认拒绝覆盖；确需改密码追加 `-reset`（会同时吊销该账号的所有会话）。

**3. 访问**

| 入口 | 地址 |
| :--- | :--- |
| 幕布首页 | http://127.0.0.1:5200/ |
| 博客列表 | http://127.0.0.1:5200/posts |
| 项目库 | http://127.0.0.1:5200/projects |
| 关于页 | http://127.0.0.1:5200/about |
| 后台登录 | http://127.0.0.1:5200/admin/login |

## 配置

**环境变量**

| 变量 | 默认 | 说明 |
| :--- | :--- | :--- |
| `SESSION_COOKIE_SECURE` | `false` | 用 HTTPS 访问时**必须设为 `true`**，否则会话 Cookie 在应用层是明文 |

**写死在 [internal/config/config.go](internal/config/config.go) 的默认值**

| 分组 | 项 | 默认 |
| :--- | :--- | :--- |
| 服务 | `ListenAddr` | `0.0.0.0:5200` |
| 服务 | `ReadHeaderTimeout` / `ReadTimeout` / `WriteTimeout` / `IdleTimeout` | `5s` / `120s` / `120s` / `120s` |
| 服务 | `MaxHeaderBytes` | `1 MiB` |
| 数据 | `DataDir` / `DBPath` | `./data` / `./data/blog.db` |
| 会话 | `SessionCookieName` / `SessionTTL` / `SessionTokenBytes` | `session` / `7d` / `32` |
| 限流 | `LoginRatePerMin` / `LoginRateWindow` | `5` 次 / `60s` |
| 密码 | `PBKDF2Iterations` / `PBKDF2KeyLength` / `SaltLength` | `600000` / `32` / `16` |
| 上传 | `MaxUploadBytes` / `UploadMemoryBytes` / `UploadsDir` | `60 MiB` / `5 MiB` / `./data/uploads` |

## 接口一览

> 鉴权约定：页面路由未登录 → `303` 跳登录页；API 路由未登录 → `401` JSON。

### 公开路由（无需登录）

| 方法 | 路径 | 说明 |
| :--- | :--- | :--- |
| GET | `/` | 幕布首页（一次进入动效） |
| GET | `/posts` | 博客列表（仅已发布，按发布时间倒序） |
| GET | `/post/{slug}` | 文章详情（草稿 / 不存在 → `404`） |
| GET | `/projects` | 项目库列表（置顶 → 排序值 → 创建时间） |
| GET | `/projects/{id}/download` | 下载项目附件（`Attachment; filename="<slug>.zip"`；无附件 → `404`） |
| GET | `/about` | 关于页 |
| GET | `/healthz` | 存活探针（不查库、不鉴权，返回 `200`） |
| GET | `/static/` | 静态资源（只暴露 `static/` 子树） |
| GET / POST | `/admin/login` | 登录页 / 提交登录 |
| POST | `/admin/logout` | 退出（无需登录也能清 Cookie） |
| POST | `/api/admin/login` | 登录（JSON 接口） |

### 管理 · 文章（均需登录）

| 方法 | 路径 | 说明 |
| :--- | :--- | :--- |
| GET | `/admin` | 后台首页 |
| GET | `/admin/posts` | 文章列表（含草稿） |
| GET | `/admin/posts/new` | 新建文章页 |
| POST | `/admin/posts` | 创建文章 |
| GET | `/admin/posts/{id}/edit` | 编辑文章页 |
| POST | `/admin/posts/{id}` | 更新文章 |
| DELETE | `/admin/posts/{id}` | 删除文章（API 鉴权，前端 `fetch` 调用） |

### 管理 · 项目库（均需登录）

| 方法 | 路径 | 说明 |
| :--- | :--- | :--- |
| GET | `/admin/projects` | 项目列表 |
| GET | `/admin/projects/new` | 新建项目页 |
| POST | `/admin/projects` | 创建项目（`multipart/form-data`） |
| GET | `/admin/projects/{id}/edit` | 编辑项目页 |
| POST | `/admin/projects/{id}` | 更新项目（可替换附件） |
| DELETE | `/admin/projects/{id}` | 删除项目（API 鉴权，同时删除附件文件） |

### 管理 · 关于页（均需登录）

| 方法 | 路径 | 说明 |
| :--- | :--- | :--- |
| GET | `/about/edit` | 编辑站主简介 |
| POST | `/about/edit` | 保存站主简介 |

### 项目表单字段（`POST /admin/projects`）

`multipart/form-data`，文本字段如下：

| 字段 | 必填 | 校验 |
| :--- | :--- | :--- |
| `title` | 是 | 非空，长度 ≤ 250 |
| `slug` | 是 | 非空，全局唯一（重复 → `409`） |
| `summary` | 是 | 非空 |
| `github_url` | 否 | 留空放行；填了必须以 `https://github.com/` 开头 |
| `demo_url` | 否 | 留空放行；填了必须以 `https://` 开头 |
| `cover_image` | 否 | 封面图 URL |
| `tech_stack` | 否 | 逗号分隔的标签串 |
| `status` | 是 | 仅 `active` / `archived` |
| `featured` | 否 | 勾选即置顶 |
| `sort_order` | 否 | 整数，解析失败按 `0` |
| `archive` | 否 | 附件文件（见下） |

### 状态码约定

| 码 | 场景 |
| :--- | :--- |
| `400` | 表单校验失败 / 路径变量非法（如 `id` 非整数） |
| `401` | API 路由未登录 |
| `404` | 资源不存在（含 `projects.ErrNotFound`） |
| `409` | slug 重复（`projects.ErrDuplicate`） |
| `413` | 请求体超过 `MaxUploadBytes` |
| `303` | 页面路由未登录 → 跳转登录页 |

## 上传与附件

- 请求体上限：`60 MiB`（`http.MaxBytesReader`），超出 → `413`；读请求体的总时长受 `ReadTimeout`（`120s`）约束。
- 仅接受 **ZIP**：读取前 4 字节做魔数校验，必须是 `50 4B 03 04`（`PK\x03\x04`）。
- 落盘文件名：16 字节随机数生成，避免用户可控路径。
- 存储位置：`UploadsDir`（默认 `./data/uploads`），数据库只存文件名。
- 下载：`GET /projects/{id}/download`，响应头带 `Attachment`，由 `http.ServeFile` 发送。
- 项目被删除 / 附件被替换时，旧文件会被清理。

## 数据模型

SQLite，单文件，迁移靠 `PRAGMA user_version` 幂等推进（`schema.sql` 编译期内嵌）。

| 表 | 用途 | 关键约束 |
| :--- | :--- | :--- |
| `posts` | 文章 | `slug` 唯一；`status` 默认 `draft`；`render_version` 记录解析器版本 |
| `projects` | 项目库 | `slug` 唯一；`status` 默认 `active`；`archive_path` 存附件文件名 |
| `admins` | 管理员 | `username` 唯一；存 `password_hash` / `salt` / `iterations` |
| `sessions` | 会话 | `token` 主键；`admin_id` 外键级联删除；`expires_at` 建索引 |
| `site_profile` | 站主简介 | 单行表（`CHECK (id = 1)`） |

## 安全设计

- **后台不裸露公网**：生产环境可在 Nginx 用 `location ^~ /admin` 只放行 Tailscale 网段，Go 层双中间件作第二道防线。
- **双中间件**：页面路由未登录 → `303` 跳登录页；fetch 数据路由未登录 → `401` JSON（避免 fetch 跟随重定向被误判为成功）。
- **XSS 防护**：CSP 禁内联（`script-src 'self'`）；Markdown 输出文本强制 HTML 转义；链接 / 图片 URL 走 scheme 白名单。
- **响应头**：[middleware.go](internal/server/middleware.go) 统一注入 `X-Content-Type-Options: nosniff`、`X-Frame-Options: DENY`、`Content-Security-Policy`、`Referrer-Policy`；`/admin` 额外 `Cache-Control: no-store`。
- **Cookie**：`HttpOnly` + `SameSite=Lax`，HTTPS 下 `Secure=true`。
- **登录限流**：防 PBKDF2 被当作 CPU 放大点打满，同时抑制暴力猜密。

## Docker 部署

镜像采用多阶段构建，运行阶段基于 `alpine`，以**非 root** 用户运行；模板与静态资源已在编译期内嵌，运行阶段无需拷贝 `web/`。

```bash
# 在仓库根目录构建（-f 指向子目录，context 仍是仓库根）
docker build \
  --build-arg UID=$(id -u) --build-arg GID=$(id -g) \
  -f docker/Dockerfile -t <registry>/homelab-blog:<version> .
```

| 项 | 值 |
| :--- | :--- |
| Builder | `golang:1.25-alpine`（`CGO_ENABLED=0`，`-ldflags="-s -w"`） |
| Runtime | `alpine:3.24` + `ca-certificates` |
| 运行用户 | `appuser`（UID/GID 由 `--build-arg` 决定，默认 `1000`） |
| 工作目录 | `/app` |
| 端口 | `5200` |
| 入口 | `./homelab-blog` |

`docker/docker-compose.yaml` 关键项：

| 字段 | 说明 |
| :--- | :--- |
| `ports` | 形如 `<本机地址>:5200:5200`，绑定到具体网卡地址 |
| `volumes` | `../data:/app/data`（相对 compose 文件所在目录解析） |
| `environment` | `SESSION_COOKIE_SECURE` |
| `restart` | `unless-stopped` |

**部署注意**

- 建议为镜像打**版本 tag**（如 `:1.0.0`）而非 `:latest`，便于回滚。
- 挂载宿主机目录要**先手工创建**，并保证其属主数字与容器运行用户一致；否则容器内进程将无写权限。
- 用 HTTPS / 反代对外时，务必把 `SESSION_COOKIE_SECURE` 设为 `true`。

## 相关文档

- [技术文档.md](技术文档.md)——完整设计：系统架构、路由表、数据模型、解析器铁律、部署方案、验收清单、决策记录（ADR）
