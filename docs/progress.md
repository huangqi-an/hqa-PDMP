# 开发进度与问题记录

> 本文用于记录项目从环境搭建到各服务开发过程中的关键决策、已完成内容和遇到的问题。架构设计见 [development.md](./development.md)，环境安装见 [setup.md](./setup.md)。

最后更新：2026-09-06

## 1. 当前进度

| 里程碑 | 内容 | 状态 |
| --- | --- | --- |
| M0 | monorepo、docker-compose、基础目录 | 已完成 |
| M1 | auth-service：注册、登录、刷新、资料管理 | 已完成 |
| M2 | vault-service：API Key CRUD、加密、软删除、reveal | 已完成 |
| M3 | web-console 前端 | 已完成 |
| M4 | 双服务 JWT 联调、Docker 打包 | 进行中 |

当前 M0-M3 的本地 MVP 已经打通：auth-service、vault-service、web-console 可以通过本地端口和 Vite 代理协作。M4 中剩余工作主要是生产部署、Dockerfile 与完整 `docker compose up`。

## 2. 已完成内容

### 2.1 开发环境

- Docker Engine 与 Docker Compose v2 已安装；
- PostgreSQL 16 已通过 `docker compose up -d postgres` 启动；
- Docker 镜像加速已配置；
- pnpm 已启用，auth-service 使用 pnpm 管理依赖。

### 2.2 auth-service

技术栈：Node.js + Express 5 + TypeScript + Prisma 7 + PostgreSQL。

已实现接口：

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST | `/api/auth/register` | 注册 |
| POST | `/api/auth/login` | 登录 |
| POST | `/api/auth/refresh` | 刷新 access token，并轮换 refresh token |
| POST | `/api/auth/logout` | 撤销 refresh token |
| GET | `/api/users/me` | 获取当前用户 |
| PATCH | `/api/users/me` | 修改当前用户资料 |

已实现能力：

- 密码使用 bcrypt 哈希存储；
- access token 无状态，refresh token 持久化到 PostgreSQL；
- refresh token 服务端只保存 SHA-256 哈希；
- refresh token 每次刷新时旧记录撤销、新记录写入；
- 统一响应格式和全局错误处理；
- zod 参数校验；
- Prisma Migrate 管理数据库迁移。

### 2.3 vault-service

技术栈：Go + Gin + GORM + golang-migrate + PostgreSQL。

已实现接口：

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/health` | 健康检查，包含数据库 ping |
| GET | `/api/keys` | 分页查询，支持 `q`、`provider` 筛选 |
| POST | `/api/keys` | 新建 API Key |
| GET | `/api/keys/:id` | 查看 API Key 元信息，不返回明文 |
| PATCH | `/api/keys/:id` | 更新 API Key 元信息或密钥值 |
| DELETE | `/api/keys/:id` | 软删除 |
| POST | `/api/keys/:id/reveal` | 查看明文，仅限所属用户 |

已实现能力：

- 使用 GORM 操作 PostgreSQL，repository 封装数据访问；
- 使用 golang-migrate 管理 `api_keys` 表；
- 使用 AES-256-GCM 加密明文，存储格式为 `base64(nonce || ciphertext)`；
- 每条记录使用随机 nonce；
- `key_hint` 只保存明文末尾 4 位；
- 列表和详情返回 `maskedKey`，不返回明文或密文；
- `user_id` 使用 TEXT，与 auth-service 的 CUID 用户 ID 对齐；
- JWT 中间件使用 HS256 校验 access token，并从 `sub` 注入 `userID`；
- 所有查询都同时限制 `id` 与 `user_id`，防止越权访问；
- GORM 软删除通过 `gorm.DeletedAt` 实现。

### 2.4 web-console

技术栈：Vue 3 + Vite + TypeScript + Vue Router + Pinia + Axios。

已实现页面：

| 路由 | 页面 | 说明 |
| --- | --- | --- |
| `/login` | 登录页 | 邮箱、密码登录 |
| `/register` | 注册页 | 邮箱、密码注册 |
| `/keys` | API Key 列表 | 列表、搜索、筛选、创建、编辑、删除、查看明文、分页 |
| `/profile` | 个人资料 | 查看当前用户、修改昵称 |

已实现能力：

- 根目录 `pnpm-workspace.yaml` 管理 `apps/*`、`services/*`、`packages/*`；
- Vite 代理将 `/api/auth`、`/api/users` 转发到 auth-service，将 `/api/keys` 转发到 vault-service；
- Axios 请求拦截器自动注入 access token；
- Axios 响应拦截器在 401 时使用 refresh token 刷新并重试原请求；
- Pinia 管理当前用户、access token 和 refresh token；
- Vue Router 守卫保护 `/keys`、`/profile`，未登录跳转 `/login`；
- `DefaultLayout` 提供控制台侧边栏和退出登录；
- API Key 明文只在用户点击“查看明文”后加载到页面内存，不做持久化。

## 3. 关键决策

| 主题 | 决策 | 原因 |
| --- | --- | --- |
| 前端 | Vue 3 + Vite + TypeScript | 熟悉且适合控制台类应用 |
| auth-service | Node + Express + TypeScript + Prisma 7 | 练习 Node，同时适合认证类高频迭代 |
| vault-service | Go + Gin + GORM + golang-migrate | 练习 Go，适合敏感数据、加密与文件类服务 |
| 包管理器 | pnpm | 安装快、锁文件可靠，适合 monorepo |
| Prisma 版本 | 固定 Prisma 7 | 8 仍为 RC，依赖链不稳定 |
| TypeScript 控制器 | 传统函数控制器 | 贴合原生 Express，避免提前引入装饰器框架 |
| Redis | 暂不引入 | 当前 refresh token 持久化用 PostgreSQL 足够 |
| refresh token | 服务端持久化并轮换 | 支持撤销、退出登录和 token 失效 |
| token 存储 | 只存 SHA-256 哈希 | 避免数据库泄露时直接暴露原始 token |
| schema 隔离 | 暂缓，先用 public | 降低首期复杂度，待 vault-service 跑通后再统一拆分 |
| vault-service 用户 ID 类型 | `api_keys.user_id` 使用 TEXT | auth-service 实际使用 CUID 字符串，不是 UUID |
| vault-service 数据库迁移 | golang-migrate | 与 Prisma Migrate 分离，避免互相覆盖 |
| vault-service 密钥存储 | 只存 AES-256-GCM 密文 | 保护模型 API Key 明文 |
| vault-service JWT 算法 | HS256 | auth-service 使用共享密钥签发 HS256 JWT |
| 前端 workspace | 根目录统一 pnpm workspace | 统一管理前端、服务和后续共享包 |
| 前端请求封装 | Axios + 拦截器 | 自动注入 token，401 时自动刷新并重试 |
| token 前端存储 | localStorage | 当前本地 MVP 简单可靠，后续可评估更安全的存储方案 |
| 前端页面结构 | 登录/注册独立页面 + 控制台布局 | 符合个人数据控制台定位 |

## 4. 已遇到的问题与解决办法

### 4.1 Docker 与数据库

**问题：`docker compose up -d postgres` 报 permission denied**

原因：当前用户不在 `docker` 组，无法访问 `/var/run/docker.sock`。

解决：

```bash
sudo usermod -aG docker $USER
newgrp docker
```

**问题：拉取 `postgres:16` 时连接 Docker Hub 失败**

原因：默认 registry 访问不稳定。

解决：在 `/etc/docker/daemon.json` 中配置阿里云镜像加速，并重启 Docker。

### 4.2 pnpm 与 Prisma

**问题：安装 Prisma 时出现 `No matching version found for @distilled.cloud/aws`**

原因：registry 拉到了 `prisma@8.0.0-rc.x`，其依赖链存在不可用的 RC 包。

解决：固定使用 Prisma 7：

```bash
pnpm add -D prisma@7.10.0
pnpm add @prisma/client@7.10.0 @prisma/adapter-pg@7.10.0 pg
```

**问题：`pnpm exec tsc --init` 提示 `Command "tsc" not found`**

原因：前一次依赖安装被 Prisma 安装失败中断，TypeScript 未真正写入 `node_modules`。

解决：先修复 Prisma 版本并重新执行 `pnpm install`，再执行 `pnpm exec tsc --init`。

**问题：`@types/bcryptjs` 安装时提示是 stub**

原因：`bcryptjs` 3.x 已自带类型声明。

解决：移除 `@types/bcryptjs`，只保留 `bcryptjs`。

**问题：Prisma 7 初始化生成的文件和旧教程不一致**

现象：生成 `prisma7.config.ts`，`schema.prisma` 的 `datasource` 中没有 `url`。

解决：

- 连接信息放在 `prisma7.config.ts`；
- generator 使用 `prisma-client`；
- 从生成的 `src/generated/prisma/client` 导入 `PrismaClient`；
- 使用 `@prisma/adapter-pg` 构造客户端。

**问题：执行 `migrate dev` 后 `prisma.refreshToken` 不存在**

原因：Prisma 7 的 `migrate dev` 不再自动执行 `prisma generate`。

解决：

```bash
pnpm exec prisma migrate dev --name <name>
pnpm exec prisma generate
```

### 4.3 Express 与 TypeScript

**问题：`app` 类型无法推断并导出**

原因：`express()` 的推断类型引用了 `@types/express-serve-static-core` 的内部类型，导出时需要显式类型。

解决：

```ts
import express, { type Express } from 'express'

const app: Express = express()
```

**问题：`jsonwebtoken` 的 `expiresIn` 报类型错误**

原因：`expiresIn` 类型不是普通 `string`，而是 `number | StringValue | undefined`；`process.env` 又是 `string | undefined`。

解决：

- secret 先判空；
- `expiresIn` 使用 `SignOptions['expiresIn']`。

```ts
import type { SignOptions } from 'jsonwebtoken'

const expiresIn: SignOptions['expiresIn'] = '15m'
```

**问题：错误中间件命中 `ZodError` 或 `AppError` 后仍然继续执行，导致重复响应**

原因：`if` 分支中没有 `return`。

解决：每个分支处理完响应后立即 `return`。

### 4.4 认证与 token

**问题：登录接口曾返回完整 `user`，包含 `passwordHash`**

原因：`prisma.user.findUnique` 默认查询全部字段。

解决：使用 `select` 明确查询 `id`、`email`、`passwordHash`，返回给客户端时只保留 `id` 和 `email`。

**问题：refresh token 最初不持久化，无法撤销**

原因：最初只生成 JWT 并返回客户端。

解决：新增 `refresh_tokens` 表，保存 refresh token 的 SHA-256 哈希；刷新时轮换，退出时撤销。

**问题：迁移中 `RefreshToken` 字段写错**

现象：`revokedAt` 被定义为非空，`createdAt` 被写成 `creatdAt`。

解决：修正 schema 后新增修复迁移，并执行：

```bash
pnpm exec prisma migrate dev --name fix_refresh_token_fields
pnpm exec prisma generate
```

**问题：`.env` 中 `JWT_REFRESH_EXPIRES_IN` 被写成 `JWT_REFLESH_EXPIRES_IN`**

解决：统一为 `JWT_REFRESH_EXPIRES_IN`，并移除等号两侧多余空格。

**问题：`.http` 测试文件曾包含真实 JWT**

原因：调试时把真实 token 写入了测试文件。

解决：真实 token 不提交，测试文件使用变量或空占位符。

### 4.5 Go 与 vault-service

**问题：`migrate` 命令找不到**

原因：`go install` 安装后的二进制默认在 `$GOPATH/bin`，而当前终端 `PATH` 没有包含该目录。

解决：

```bash
export PATH="$PATH:/usr/local/go/bin:$(go env GOPATH)/bin"
```

或将上述内容加入 `~/.bashrc`，然后重新打开终端。

**问题：`migrate -database "$DATABASE_URL" -path migrations up` 报 `pq: SSL is not enabled on the server`**

原因：本地 PostgreSQL 没有启用 SSL，而 Go PostgreSQL 驱动默认尝试 SSL。

解决：本地 `DATABASE_URL` 使用：

```text
postgresql://hqa:hqa@localhost:5432/hqa_pdmp?sslmode=disable
```

**问题：`api_keys.user_id` 最初定义为 UUID，但 auth-service 用户 ID 实际是 CUID**

原因：auth-service 的 `User.id` 使用 `String @default(cuid())`，而 vault-service 首版迁移把 `user_id` 写成了 UUID。

解决：保持 auth-service 不变，新增 `000002_change_api_keys_user_id_to_text` 迁移，将 `api_keys.user_id` 改为 TEXT，并在 GORM 模型中使用 `string`。

**问题：JWT 中间件最初使用 ES256 校验，导致所有合法 token 都被拒绝**

原因：auth-service 使用 HS256 签名，但 vault-service 中间件写成了 `jwt.SigningMethodES256`。

解决：改为 `jwt.SigningMethodHS256`，并使用 `[]byte(JWT_ACCESS_SECRET)` 作为校验密钥。

**问题：PATCH 路由写成 `/id`，不是 `/:id`**

现象：前端请求 `/api/keys/:id` 无法匹配更新接口。

解决：统一为 `/api/keys/:id`。

**问题：handler 的 `handleServiceError` 遇到未知错误时不返回响应**

现象：客户端请求可能一直等待，直到超时。

解决：非 `ErrAPIKeyNotFound` 错误记录日志并返回 `500 / 9999`。

**问题：Go 测试时默认 GOCACHE 位于只读目录**

现象：部分环境执行 `go test ./...` 会报 `read-only file system`。

解决：

```bash
GOCACHE=/tmp/hqa-gocache go test ./...
```

### 4.6 Vue 与 web-console

**问题：前端请求 `/api/users` 时 Vite 没有转发到 auth-service**

原因：最初的代理只配置了 `/api/auth` 和 `/api/keys`，遗漏了个人资料接口。

解决：在 `vite.config.ts` 中补充：

```ts
"/api/users": "http://localhost:3001",
```

**问题：`http.ts` 中刷新 token 结果被推断为 `string`**

现象：TypeScript 报 `Property 'accessToken' does not exist on type 'string'`。

原因：`refreshPromise` 被错误地声明为 `Promise<string | null>`，而实际刷新接口返回 `{ accessToken, refreshToken }`。

解决：定义 `RefreshSession` 类型，并为刷新请求补充 Axios 泛型。

**问题：`router/index.ts` 中 store import 带了 `.ts` 后缀**

解决：统一改为：

```ts
import { useAuthStore } from "../stores/auth";
```

**问题：根级 workspace 未建立，`services/auth-service` 内部存在独立 workspace 文件**

原因：auth-service 初始化时把 `pnpm-workspace.yaml` 放在了服务目录内。

解决：在项目根目录建立统一的 `pnpm-workspace.yaml`，包含 `apps/*`、`services/*`、`packages/*`，并移除 auth-service 内部的独立 workspace 文件。

## 5. 当前环境变量约定

### auth-service

```text
PORT
DATABASE_URL
JWT_ACCESS_SECRET
JWT_REFRESH_SECRET
JWT_ACCESS_EXPIRES_IN
JWT_REFRESH_EXPIRES_IN
```

### vault-service

```text
PORT
DATABASE_URL
JWT_ACCESS_SECRET
ENCRYPTION_KEY
```

## 6. 下一步计划

1. 完善前端体验：创建/编辑表单互斥、删除后页码回退、展示后端业务错误、接入 UI 组件库；
2. 完善 M4：为 auth-service、vault-service 和 web-console 编写 Dockerfile；
3. 扩展 `docker-compose.yml`，支持完整的一键部署；
4. 部署阶段由 Nginx 或 Caddy 统一代理 `/api/auth`、`/api/users`、`/api/keys` 和前端静态资源；
5. 根据联调结果继续统一错误码、日志和 API 契约。
