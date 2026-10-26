# SMS 转发服务器

这是一个基于 Go 的短信接收服务，支持通过 HTTP webhook 接收短信，自动提取验证码，并提供历史记录查询。默认子路径采用 `/sms`，可通过 Nginx 反向代理转发到该服务。

## 功能概览

- 接收短信 webhook：`POST /sms/webhook`
- 健康检查：`GET /sms/healthz`
- 历史查询：`GET /sms/history?limit=20&page=1`
- 自动提取短信验证码（4~8 位数字）
- 结构化日志：request_id、method、path、status、latency、error
- SQLite 持久化
- Docker 支持与 Nginx 子路径转发示例

## 目录结构

```text
sms_server
├── cmd
│   └── server
│       └── main.go
├── deploy
│   ├── docker
│   │   ├── deploy.sh
│   │   ├── docker-compose.prod.yml
│   │   └── rollback.sh
│   └── nginx
│       └── sms_subpath.conf
├── scripts
│   ├── deploy.bat
│   ├── deploy.ps1
│   └── deploy_remote.ps1
├── internal
│   ├── config
│   │   └── config.go
│   ├── handler
│   │   └── handlers.go
│   ├── logging
│   │   └── logger.go
│   ├── middleware
│   │   └── requestlog.go
│   ├── service
│   │   └── sms.go
│   └── store
│       └── store.go
├── Dockerfile
├── go.mod
├── .env.example
├── README.md
└── data
```

## 本地运行

```bash
go mod download
mkdir -p data
PORT=8322 HOST=0.0.0.0 DATA_DIR=./data LOG_LEVEL=info go run ./cmd/server
```

访问：

- 健康检查：`http://127.0.0.1:8322/sms/healthz`
- 历史记录：`http://127.0.0.1:8322/sms/history?limit=20&page=1`

## API 说明

### 1) 接收短信 webhook

```bash
curl -X POST http://127.0.0.1:8322/sms/webhook \
  -H 'Content-Type: application/json' \
  -d '{
    "from": "10086",
    "contact_name": "中国移动",
    "phone_area": "86",
    "sms": "您的验证码是123456，请在5分钟内完成验证。",
    "sim_slot": "SIM1(13800138000)",
    "sim_sub_id": "1",
    "device_name": "DemoPhone",
    "receive_time": "2025-03-15T14:30:00Z"
  }'
```

返回示例：

```json
{
  "result": "ok",
  "code": "SUCCESS",
  "message": "短信接收成功",
  "data": {
    "id": 1,
    "extracted_code": "123456",
    "phone_number": "13800138000"
  }
}
```

### 2) 健康检查

```bash
curl http://127.0.0.1:8322/sms/healthz
```

### 3) 历史查询

```bash
curl 'http://127.0.0.1:8322/sms/history?limit=20&page=1'
```

## 兼容性说明

为了确保与旧服务的兼容性，本服务同时保留旧路由：

- `POST /v1/sms/receive`
- `GET /v1/sms/history`
- `GET /v1/sms/healthz`

其中 `/sms` 是推荐路径，适合 Nginx 子路径代理。 

## 环境变量

```bash
PORT=8322
HOST=0.0.0.0
DATA_DIR=./data
LOG_LEVEL=info
GIN_MODE=release
```

## Docker 部署

### 构建镜像

```bash
docker build -t sms-server-go .
```

### 运行容器

```bash
docker run -d --name sms-server-go \
  -p 8322:8322 \
  -v $(pwd)/data:/app/data \
  sms-server-go
```

## 自动化部署（Docker，参考 financial_prediction 脚本风格）

新增脚本：

- `scripts/deploy.ps1`
- `scripts/deploy.bat`（Windows 入口）
- `scripts/deploy_remote.ps1`（本地打包 + 远程上传 + 远程部署）
- `deploy/docker/deploy.sh`（服务器执行部署，含健康检查与失败回滚）
- `deploy/docker/rollback.sh`（手动回滚）

脚本流程：

1. `go test ./...`
2. `docker build` 构建镜像
3. `docker save` 导出镜像并打包部署材料
4. 生成 `build/sms-server-deploy.tar.gz`
5. 可选：通过 `checksum + HMAC-SHA256 signature` 上传部署包

### 本地生成部署包

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\deploy.ps1 -ImageTag v1.0.0 -SkipUpload
```

### 上传部署包（可选）

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\deploy.ps1 `
  -ImageTag v1.0.0 `
  -UploadUrl "https://your-domain/api/upload" `
  -SecretKey "replace-with-your-secret"
```

### 服务器侧部署

```bash
mkdir -p /opt/sms-server && cd /opt/sms-server
tar -xzf sms-server-deploy.tar.gz
chmod +x deploy/docker/deploy.sh deploy/docker/rollback.sh
cp .env.example .env
mkdir -p data
./deploy/docker/deploy.sh --image sms-server-go:v1.0.0
```

说明：

- `deploy/docker/docker-compose.prod.yml` 默认使用 `.env` 中的 `PORT/HOST/DATA_DIR/LOG_LEVEL/GIN_MODE`。
- 如需 Nginx 子路径代理，可使用 `deploy/nginx/sms_subpath.conf`。
- `deploy/docker/deploy.sh` 会执行健康检查（默认 `http://127.0.0.1:8322/sms/healthz`，超时 60 秒），失败时自动回滚到部署前镜像。

### 一键远程发布（Windows）

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\deploy_remote.ps1 `
  -RemoteHost 203.0.113.10 `
  -RemoteUser root `
  -RemotePath /opt/sms-server `
  -ImageTag v1.0.0
```

可选参数：

- `-IdentityFile C:\Users\you\.ssh\id_rsa`
- `-SshPort 22`
- `-HealthUrl http://127.0.0.1:8322/sms/healthz`
- `-HealthTimeout 90`
- `-SkipTests`

### 手动回滚

```bash
cd /opt/sms-server
./deploy/docker/rollback.sh
```

## Nginx 子路径转发示例

配置文件：`deploy/nginx/sms_subpath.conf`

```nginx
server {
    listen 80;
    server_name sms.example.com;

    location /sms/ {
        proxy_pass http://127.0.0.1:8322/sms/;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header X-Request-ID $request_id;
        proxy_connect_timeout 5s;
        proxy_read_timeout 30s;
        proxy_send_timeout 30s;
    }

    location / {
        return 404;
    }
}
```

## 测试

```bash
go test ./...
```

说明：

- `internal/handler` 已补充基于 SmsForwarder webhook 字段链路（`/v1/sms/receive -> /sms/history`）的单元测试。
- 在未安装 C 编译器（如 `gcc`）的环境中，SQLite 相关 handler 单测会自动 `skip`，其余测试照常运行。

## 变更说明

本版本已从 Python/FastAPI 重构为 Go 服务，目标是：

- 更快启动与更低资源占用
- 更容易部署在容器和 VPS 环境
- 更清晰的日志结构与更稳定的 API 兼容性
- 统一使用 `/sms` 子路径承接功能

## 许可证

MIT
