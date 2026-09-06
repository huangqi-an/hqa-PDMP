# ---------- 前端构建阶段 ----------
FROM node:24-alpine AS builder

RUN corepack enable \
  && corepack prepare pnpm@11.25.0 --activate

WORKDIR /workspace

# 和 auth-service 一样，构建上下文是项目根目录
COPY pnpm-workspace.yaml pnpm-lock.yaml ./

# 先复制 package.json，利用依赖缓存
COPY apps/web-console/package.json apps/web-console/package.json

# 配置国内 npm registry
RUN pnpm config set registry https://registry.npmmirror.com

# 只安装 web-console 依赖
RUN pnpm install --filter web-console --frozen-lockfile

# 复制前端源码
COPY apps/web-console apps/web-console

# 执行生产构建
RUN pnpm --filter web-console build

# ---------- 运行阶段 ----------
FROM nginx:1.29-alpine

COPY apps/web-console/nginx.conf /etc/nginx/conf.d/default.conf

COPY --from=builder /workspace/apps/web-console/dist /usr/share/nginx/html

EXPOSE 80