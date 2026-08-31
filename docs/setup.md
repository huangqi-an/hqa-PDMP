# 开发环境搭建

本文记录本地开发环境的安装与启动步骤，属于一次性配置。业务架构与接口设计见 [development.md](./development.md)。

## 1. 前置要求

- Node.js 20+ 与 npm（可选 pnpm）；
- Go 1.22+；
- Docker Engine + Docker Compose v2。

## 2. 安装 Docker

Ubuntu 下安装 Docker Engine 与 compose 插件，参考官方步骤，关键命令如下：

```bash
sudo apt update
sudo apt install -y ca-certificates curl gnupg
sudo install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
sudo chmod a+r /etc/apt/keyrings/docker.gpg
echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu $(. /etc/os-release && echo $VERSION_CODENAME) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
sudo apt update
sudo apt install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
sudo usermod -aG docker $USER
```

加入 docker 组后需重新登录终端，或执行 `newgrp docker`，否则访问 `/var/run/docker.sock` 会报 permission denied。

验证：

```bash
docker --version
docker compose version
```

## 3. 配置国内镜像加速（可选）

如果 `docker pull` 连接 `registry-1.docker.io` 失败，可在 `/etc/docker/daemon.json` 配置镜像加速：

```json
{
  "registry-mirrors": [
    "https://<你的阿里云ID>.mirror.aliyuncs.com"
  ]
}
```

阿里云个人加速地址在 `cr.console.aliyun.com` 的「镜像工具 → 镜像加速器」获取。注意这是个人账号信息，不要提交到仓库。

修改后重启并验证：

```bash
sudo systemctl daemon-reload
sudo systemctl restart docker
docker info | grep -A 10 "Registry Mirrors"
```

## 4. 启动 PostgreSQL

项目根目录的 `docker-compose.yml` 提供 `postgres:16` 服务：

```bash
docker compose up -d postgres
docker compose ps
docker exec hqa-postgres pg_isready -U hqa
```

`pg_isready` 输出 `accepting connections` 即表示数据库可用。

## 5. 连接信息

本地开发默认值如下，生产环境需替换：

```text
host:      localhost
port:      5432
user:      hqa
password:  hqa
database:  hqa_pdmp
```

Prisma 使用的 `DATABASE_URL`：

```text
postgresql://hqa:hqa@localhost:5432/hqa_pdmp?schema=public
```

## 6. 各服务启动方式

端口约定、Vite 代理与批量启动方式见 [development.md](./development.md) 第 10 节。
