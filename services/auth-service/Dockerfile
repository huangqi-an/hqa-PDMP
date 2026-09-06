FROM node:24-alpine

# 启用 Node 自带的 corepack，并固定使用和本地一致的 pnpm 版本
RUN corepack enable \
  && corepack prepare pnpm@11.25.0 --activate

# 构建上下文是项目根目录，所以这里先进入 workspace 根
WORKDIR /workspace

# 先复制 pnpm workspace 和锁文件，让依赖安装可以命中缓存
COPY pnpm-workspace.yaml pnpm-lock.yaml ./

# 复制 auth-service 自己的 package 配置
COPY services/auth-service/package.json services/auth-service/package.json

# 复制 Prisma schema、迁移和 Prisma 配置
COPY services/auth-service/prisma services/auth-service/prisma
COPY services/auth-service/prisma7.config.ts services/auth-service/prisma7.config.ts

# 复制 TypeScript 源码
COPY services/auth-service/src services/auth-service/src

# 安装 auth-service 及其 workspace 依赖
RUN pnpm config set registry https://registry.npmmirror.com
RUN pnpm install --filter auth-service --frozen-lockfile

# 生成 Prisma Client
RUN pnpm --filter auth-service exec prisma generate

# 运行时进入 auth-service 目录，方便执行 pnpm 脚本
WORKDIR /workspace/services/auth-service

EXPOSE 3001

CMD ["sh", "-c", "pnpm exec prisma migrate deploy && pnpm exec tsx src/index.ts"]