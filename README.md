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
│   └── nginx
│       └── sms_subpath.conf
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

## 变更说明

本版本已从 Python/FastAPI 重构为 Go 服务，目标是：

- 更快启动与更低资源占用
- 更容易部署在容器和 VPS 环境
- 更清晰的日志结构与更稳定的 API 兼容性
- 统一使用 `/sms` 子路径承接功能

## 许可证

MIT
