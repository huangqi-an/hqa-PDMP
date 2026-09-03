# 开发进度与问题记录

> 本文用于记录项目从环境搭建到各服务开发过程中的关键决策、已完成内容和遇到的问题。架构设计见 [development.md](./development.md)，环境安装见 [setup.md](./setup.md)。

最后更新：2026-09-04

## 1. 当前进度

| 里程碑 | 内容 | 状态 |
| --- | --- | --- |
| M0 | monorepo、docker-compose、基础目录 | 已完成 |
| M1 | auth-service：注册、登录、刷新、资料管理 | 已完成 |
| M2 | vault-service：API Key CRUD、加密、软删除、reveal | 进行中 |
| M3 | web-console 前端 | 未开始 |
| M4 | 双服务 JWT 联调、Docker 打包 | 未开始 |

当前 `vault-service` 已完成 Go module 初始化与依赖安装，尚未开始 Gin 路由和数据库迁移。

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

当前仅完成：

- `services/vault-service/go.mod`；
- `services/vault-service/go.sum`；
- 基础 Go 依赖：Gin、GORM、PostgreSQL 驱动、JWT、godotenv、google/uuid。

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

1. 在 `vault-service` 中实现 Gin `/health`；
2. 使用 golang-migrate 创建 `api_keys` 表；
3. 接入 GORM；
4. 实现 JWT 鉴权中间件；
5. 实现 AES-256-GCM 加解密；
6. 实现 API Key CRUD、软删除和 reveal；
7. 与 auth-service 做 JWT 联调。
