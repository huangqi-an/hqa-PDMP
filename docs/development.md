# hqa-PDMP 开发文档

## 1. 项目概述

hqa-PDMP 是一个个人数据管理平台。首期 MVP 聚焦两件事：

- 用户注册、登录与个人资料管理；
- 管理用户拥有的各类模型 API Key（增删改查、搜索、分页、软删除）。

后续可在此基础上扩展 AI 对话、Agent 智能体、个人网盘等能力。

项目同时用于练习两套后端技术栈：Node.js + Express 和 Go + Gin。

## 2. 技术选型

| 模块 | 选型 |
| --- | --- |
| 前端 | Vue 3 + Vite + TypeScript |
| 认证服务 auth-service | Node.js + Express + TypeScript + Prisma |
| 密钥库服务 vault-service | Go + Gin + GORM + golang-migrate |
| 数据库 | PostgreSQL |
| 认证方式 | JWT（HS256） |
| 密钥加密 | AES-256-GCM |
| 容器编排 | Docker Compose |

## 3. 服务职责划分

划分原则：

- 靠近用户、迭代频繁、与前端或第三方 SDK 耦合紧密的模块，放 Node；
- 靠近数据、高并发、长时间运行、CPU/文件/任务密集的模块，放 Go；
- 两者皆可时，优先补足你更想练习的那门语言。

现状划分：

| 功能 | 归属 | 理由 |
| --- | --- | --- |
| 用户 / 认证 / 会话 | Node（auth-service） | 贴近用户，迭代快 |
| API Key 管理 | Go（vault-service） | 敏感数据 + 加密存储 |

未来扩展：

| 功能 | 建议归属 | 理由 |
| --- | --- | --- |
| AI 对话（实时流式） | Node，可新建 chat-service 或 BFF | 流式、SSE、第三方 SDK 生态强 |
| Agent 智能体 / 任务编排 | Go，agent-service | 多步调用、高并发、长时间运行 |
| 个人网盘 / 文件存储 | Go，storage-service | 文件流、上传下载、低内存 |
| 统一网关 / 前端聚合 | Node，BFF | 面向前端的鉴权、聚合、流式转发 |

AI 对话建议分成两层：面向用户的会话管理与流式转发放 Node，真正调用模型、工具链和多步执行的引擎放 Go。这样既能练到跨服务协作，也避免 Node 承载过重。

## 4. 仓库结构与共享代码

```text
hqa-PDMP/
  apps/                      # 前端应用
    web-console/             # 主控制台（Vue 3 + Vite）
      src/
        api/                 # 接口请求封装
        views/               # 页面
        components/          # 组件
        stores/              # 状态管理
        router/              # 路由
  services/                  # 后端服务
    auth-service/            # Node + Express
      prisma/
        schema.prisma
        migrations/
      src/
        config/
        middleware/
        routes/
        services/
        utils/
    vault-service/           # Go + Gin
      cmd/server/main.go
      internal/
        config/
        handler/
        middleware/
        model/
        repository/
        service/
      migrations/
  packages/                  # 跨端共享代码
    shared/                  # 共享类型、常量、工具
    ui/                      # 共享 UI 组件
  docs/
    development.md
  docker-compose.yml
  README.md
```

### 4.1 跨端共享代码（packages）

`packages/` 用于在多个前端之间，以及前端与 Node 服务之间复用 TypeScript 代码。Go 服务无法直接 import TypeScript，跨语言共享通过 OpenAPI / JSON Schema 契约或代码生成实现，不放进 `packages/`。

`packages/shared` 建议放：

- API 契约类型：`ApiResponse<T>`、分页结构、登录/密钥 DTO、错误码枚举；
- 常量：错误码值、分页默认值、provider 枚举；
- 纯函数：密钥掩码、日期格式化、参数校验；
- 校验 schema（可选，zod），供前后端共用。

`packages/ui` 建议放：通用 Vue 组件（按钮、输入框、弹窗、表格、分页、Toast、确认框）、设计 token 与主题变量、图标封装。

使用方式：根目录 `package.json` 声明 workspaces，`packages/shared` 命名为 `@hqa/shared`，前端通过 workspace 依赖引入。

根目录 `package.json`：

```json
{
  "name": "hqa-pdmp",
  "private": true,
  "workspaces": ["apps/*", "packages/*"]
}
```

`packages/shared/package.json`：

```json
{
  "name": "@hqa/shared",
  "version": "0.0.0",
  "type": "module",
  "main": "./src/index.ts",
  "types": "./src/index.ts"
}
```

`apps/web-console/package.json` 引入：

```json
{
  "dependencies": {
    "@hqa/shared": "workspace:*"
  }
}
```

pnpm 使用 `workspace:*`，npm 使用 `"*"`。开发期让 `main`/`types` 直接指向 `src/index.ts`，Vite 与 tsx 可直接编译，无需先构建；发布前再用 tsc 或 tsup 构建出 `dist` 目录。

建议先不急于抽包，等出现第二处重复时再抽取，避免过度设计；最先值得共享的是统一响应格式与错误码。

## 5. 数据库设计

两个服务共用同一个 PostgreSQL 实例，但使用不同的 schema 隔离：

- `auth`：认证服务管理；
- `vault`：密钥库服务管理。

### 5.1 users（auth schema）

| 字段 | 类型 | 约束 |
| --- | --- | --- |
| id | UUID | 主键，默认 gen_random_uuid() |
| email | text | 唯一，非空 |
| password_hash | text | 非空 |
| created_at | timestamptz | 默认 now() |
| updated_at | timestamptz | 默认 now() |

### 5.2 api_keys（vault schema）

| 字段 | 类型 | 约束 |
| --- | --- | --- |
| id | UUID | 主键，默认 gen_random_uuid() |
| user_id | UUID | 非空，来源为认证服务的用户 id |
| provider | text | 非空 |
| name | text | 非空 |
| encrypted_key | text | 非空，base64 编码的密文 |
| key_hint | text | 非空，仅保存末尾 4 位 |
| notes | text | 可空 |
| tags | text[] | 默认空数组 |
| created_at | timestamptz | 默认 now() |
| updated_at | timestamptz | 默认 now() |
| deleted_at | timestamptz | 可空，用于软删除 |

索引建议：`(user_id, deleted_at)` 组合索引，便于按用户查询未删除的数据。

### 5.3 跨服务数据一致性

`api_keys.user_id` 与 `users.id` 之间存在逻辑外键关系。首期不建跨 schema 的外键约束，以降低服务间耦合；数据一致性由 vault 服务在鉴权后使用 JWT 中的 `sub`（用户 id）保证。

### 5.4 迁移管理

- `auth` schema 由 Prisma Migrate 管理；
- `vault` schema 由 golang-migrate 管理。

两套迁移文件分开维护，避免互相覆盖。

## 6. API 设计

统一响应格式：

```json
{
  "code": 0,
  "message": "ok",
  "data": {}
}
```

`code` 为 `0` 表示成功，非 `0` 表示业务错误。HTTP 状态码同时表达语义：成功 `2xx`，参数错误 `400`，未认证 `401`，无权限 `403`，不存在 `404`，服务器错误 `500`。

### 6.1 认证服务 auth-service

| 方法 | 路径 | 鉴权 | 说明 |
| --- | --- | --- | --- |
| POST | /api/auth/register | 否 | 注册，入参 email、password |
| POST | /api/auth/login | 否 | 登录，入参 email、password |
| POST | /api/auth/refresh | 否 | 用 refresh token 换新 access token |
| POST | /api/auth/logout | 是 | 登出 |
| GET | /api/users/me | 是 | 获取当前用户资料 |
| PATCH | /api/users/me | 是 | 修改当前用户资料 |

登录成功返回：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "accessToken": "xxx",
    "refreshToken": "yyy",
    "user": {
      "id": "uuid",
      "email": "user@example.com"
    }
  }
}
```

### 6.2 密钥库服务 vault-service

所有接口都需要携带 `Authorization: Bearer <accessToken>`。

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | /api/keys | 分页查询，支持 q、provider 筛选 |
| POST | /api/keys | 新建密钥 |
| GET | /api/keys/:id | 查看密钥元信息，不返回明文 |
| PATCH | /api/keys/:id | 修改密钥元信息或密钥值 |
| DELETE | /api/keys/:id | 软删除 |
| POST | /api/keys/:id/reveal | 返回明文密钥，仅限本人 |

列表与详情中的密钥展示形式：

```json
{
  "id": "uuid",
  "provider": "openai",
  "name": "default",
  "maskedKey": "sk-****abcd",
  "keyHint": "abcd",
  "notes": "主力 key",
  "tags": ["prod"],
  "createdAt": "2026-08-31T00:00:00Z",
  "updatedAt": "2026-08-31T00:00:00Z"
}
```

明文密钥只通过 `reveal` 接口返回，且只在用户主动查看时返回。

## 7. 认证与安全

### 7.1 密码

首期使用 bcrypt（cost 10 或 12）存储密码哈希，禁止明文或可逆加密。

### 7.2 JWT

- 签名算法：HS256；
- 两个服务共享同一个 `JWT_SECRET`；
- access token 有效期建议 15 分钟；
- refresh token 有效期建议 7 天；
- JWT 的 `sub` 为用户 id，vault 服务校验签名和有效期后信任该 id。

首期 refresh token 由客户端保存，不做服务端状态管理；后续可扩展为 httpOnly cookie 或服务端 denylist。

### 7.3 API Key 加密

- 算法：AES-256-GCM；
- 密钥：环境变量中的 32 字节主密钥；
- 每条记录使用随机 nonce；
- 存储格式：`base64(nonce || ciphertext)`；
- 禁止在日志、响应、异常信息中输出明文密钥；
- 列表和详情只返回 `keyHint` 与 `maskedKey`。

## 8. 前端设计

### 8.1 页面

| 路由 | 页面 | 说明 |
| --- | --- | --- |
| /login | 登录页 | 邮箱、密码登录 |
| /register | 注册页 | 邮箱、密码注册 |
| /keys | 密钥列表 | 列表、搜索、筛选、分页 |
| /profile | 个人资料 | 查看与修改资料 |

### 8.2 交互要点

- access token 保存在内存或 localStorage，请求拦截器自动注入；
- access token 过期时自动用 refresh token 刷新后重试原请求；
- 新建/编辑密钥使用弹窗表单；
- 删除密钥使用二次确认；
- 明文密钥只在用户点击“查看”时请求 reveal 接口，不做本地缓存。

## 9. 环境变量

### 9.1 auth-service

```text
PORT
DATABASE_URL
JWT_SECRET
```

### 9.2 vault-service

```text
PORT
DATABASE_URL
JWT_SECRET
ENCRYPTION_KEY
```

`ENCRYPTION_KEY` 建议为 32 字节随机数的 base64 编码。

## 10. 本地开发与运行

### 10.1 端口约定

本地开发时每个服务各起一个进程、各占一个端口：

| 服务 | 端口 |
| --- | --- |
| web（Vite） | 5173 |
| auth-service | 3001 |
| vault-service | 8080 |
| postgres | 5432 |

### 10.2 前端如何访问多个后端

前端统一请求同源的 `/api/...`，开发期由 Vite 代理按路径转发，避免跨域：

```ts
// vite.config.ts
server: {
  proxy: {
    '/api/auth': 'http://localhost:3001',
    '/api/keys': 'http://localhost:8080',
  },
},
```

生产环境由 Nginx（独立的反向代理软件，不是 Node.js）或 Caddy 做统一入口：

```text
浏览器 → app.example.com
          /api/auth/*  → auth-service
          /api/keys/*  → vault-service
          /            → 前端静态资源
```

端口不影响鉴权：两个服务验证同一个 `JWT_SECRET` 签发的 token，因此前端持有一个 token 即可访问任意服务。

### 10.3 批量启动

数据库用 Docker 单独启动：

```bash
docker compose up -d postgres
```

三个应用推荐用进程管理器一条命令拉起。方案一，根目录 `package.json` + concurrently：

```json
{
  "scripts": {
    "dev": "concurrently -n web-console,auth,vault -c auto \"npm --prefix apps/web-console run dev\" \"npm --prefix services/auth-service run dev\" \"cd services/vault-service && air\""
  }
}
```

```bash
npm run dev
```

方案二，Procfile + overmind/foreman：

```text
web-console: cd apps/web-console && npm run dev
auth: cd services/auth-service && npm run dev
vault: cd services/vault-service && air
```

```bash
overmind start
```

overmind 支持单服务重启、分别查看日志，适合长期使用；concurrently 更轻量，适合起步。

### 10.4 热更新

- web-console（前端）：Vite 自带热更新；
- auth-service：`tsx watch`、`node --watch` 或 nodemon；
- vault-service：`air` 自动监听文件并重新编译。

### 10.5 启动步骤

1. `docker compose up -d postgres` 启动数据库；
2. 初始化 `auth` schema 迁移并启动 auth-service；
3. 初始化 `vault` schema 迁移并启动 vault-service；
4. 启动前端开发服务器，或直接 `npm run dev` / `overmind start` 一次拉起全部应用。

`docker-compose.yml` 首期只需包含 PostgreSQL 服务，本地开发不把三个应用容器化，以免热更新变慢；部署阶段再补完整的 `docker compose up`。

## 11. 里程碑

| 阶段 | 内容 |
| --- | --- |
| M0 | 初始化 monorepo、docker-compose、基础目录 |
| M1 | Express 认证服务：注册、登录、刷新、个人资料 |
| M2 | Gin 密钥库服务：密钥 CRUD、加密、软删除、reveal |
| M3 | Vue 前端：登录、注册、密钥管理、个人资料 |
| M4 | 双服务 JWT 联调、统一错误处理、Docker 打包 |
| M5 | 后续扩展：AI 对话、Agent、个人网盘 |

## 12. 后续扩展方向

- AI 对话：接入已保存的模型 Key，提供统一对话入口；
- Agent 智能体：将密钥与工具调用、任务编排结合；
- 个人网盘：文件上传、目录管理、对象存储。

## 13. 开发约定

- TypeScript 开启严格模式，使用 ESLint + Prettier；
- Go 代码使用 gofmt，建议接入 golangci-lint；
- 所有接口使用统一响应格式和错误码；
- 日志使用结构化日志，禁止记录密码、明文密钥、JWT；
- 提交信息保持清晰，按模块描述变更。
