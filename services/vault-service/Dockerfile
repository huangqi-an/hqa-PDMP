# ---------- 编译阶段 ----------
FROM golang:1.27-alpine AS builder

# 国内 Go 模块代理，避免从 proxy.golang.org 拉取超时
ENV GOPROXY=https://goproxy.cn,direct \
  CGO_ENABLED=0 \
  GOOS=linux

WORKDIR /src

# 先复制 go.mod 和 go.sum，利用 Docker 层缓存
COPY go.mod go.sum ./

# 下载 Go 依赖
RUN go mod download

# 复制服务源码
COPY cmd ./cmd
COPY internal ./internal

# 编译 vault-service
RUN go build -o /out/vault-service ./cmd/server

# 编译 golang-migrate CLI
RUN go install -tags 'postgres' \
  github.com/golang-migrate/migrate/v4/cmd/migrate@v4.19.1

# ---------- 运行阶段 ----------
FROM alpine:3.22

RUN apk add --no-cache ca-certificates

WORKDIR /app

# 复制服务二进制
COPY --from=builder /out/vault-service ./vault-service

# 复制 migrate 命令
COPY --from=builder /go/bin/migrate ./migrate

# 复制迁移 SQL
COPY migrations ./migrations

EXPOSE 8080

CMD ["sh", "-c", "./migrate -database \"$DATABASE_URL\" -path ./migrations up && ./vault-service"]