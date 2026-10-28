# SMS 转发服务器

这是一个基于 Go 的短信接收服务，支持通过 HTTP webhook 接收短信，自动提取验证码，并提供历史记录查询。默认子路径采用 `/sms`，可通过 Nginx 反向代理转发到该服务。

前端 UI 已升级为 **React + TypeScript + Rsbuild** 架构，支持功能/设置模块拆分与移动端适配。

## 功能概览

- 接收短信 webhook：`POST /sms/webhook`
- 健康检查：`GET /sms/healthz`
- 历史查询：`GET /sms/history?limit=20&page=1`
- UI 页面：`GET /sms/ui`
- 支持上报/历史接口 key 鉴权（支持 key 轮换）
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
npm install
npm run build:ui   # 生成 frontend/static/app（该目录不入库，必须本地构建）
go mod download
mkdir -p data
PORT=8322 HOST=0.0.0.0 DATA_DIR=./data LOG_LEVEL=info go run ./cmd/server
```

访问：

- 健康检查：`http://127.0.0.1:8322/sms/healthz`
- 历史记录：`http://127.0.0.1:8322/sms/history?limit=20&page=1`
- UI 页面：`http://127.0.0.1:8322/sms/ui`

UI 页面说明：

- UI 按“功能模块 / 设置模块”拆分，设置集中在设置模块中维护。
- 页面会先检查是否已保存查询 key。未设置时，会提示先在设置模块保存 key（`localStorage` 键名：`sms_report_auth_key`）。
- 历史查询请求会自动携带保存后的 key（`X-Report-Key`），并支持分页查询。
- 页面会展示已拼接 key 的 webhook 地址与兼容地址，可一键复制用于第三方回调配置。
- 下载功能只保留标题栏按钮：移动端点击直接跳转 APK 下载；桌面端点击弹窗展示下载链接与二维码。

## API 说明

### 1) 接收短信 webhook

```bash
curl -X POST http://127.0.0.1:8322/sms/webhook \
  -H 'Content-Type: application/json' \
  -H 'X-Report-Key: <your-report-key>' \
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

兼容性说明：

- 支持 `application/json`（推荐）和 `application/x-www-form-urlencoded`（SmsForwarder 的“测试请求”常用）两种请求体。
- 必填字段仍为 `from` 与 `sms`（表单模式下也可兼容 `sender`/`phone` 和 `msg`/`message`/`content`）。

> 若未配置 `REPORT_AUTH_KEY`，可不传 `X-Report-Key`；若配置了该变量，则上报接口必须携带正确密钥。

### 2) 健康检查

```bash
curl http://127.0.0.1:8322/sms/healthz
```

### 3) 历史查询

```bash
curl 'http://127.0.0.1:8322/sms/history?limit=20&page=1' \
  -H 'X-Report-Key: <your-report-key>'
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
REPORT_AUTH_KEY=
REPORT_AUTH_KEYS=
```

说明：

- 当 `REPORT_AUTH_KEY` 与 `REPORT_AUTH_KEYS` 都为空时，上报/历史接口不做鉴权（兼容旧客户端）。
- 当配置了任一 key 后，以下接口必须携带正确 key：
  - `POST /sms/webhook`
  - `POST /v1/sms/receive`
  - `GET /sms/history`
  - `GET /v1/sms/history`
- 可通过 `REPORT_AUTH_KEYS` 配置多个 key（逗号分隔）用于轮换，`REPORT_AUTH_KEY` 也会自动并入可用 key 集合。
- 请求可通过以下任一方式传递 key：
  - `X-Report-Key: <key>`，或
  - `Authorization: Bearer <key>`，或
  - URL 查询参数：`?report_key=<key>`（适合仅支持 URL 的第三方上报）

## Docker 部署

### 构建镜像

```bash
docker build \
  --build-arg GO_BASE_IMAGE=docker.m.daocloud.io/library/golang:1.22-alpine \
  --build-arg RUNTIME_BASE_IMAGE=docker.m.daocloud.io/library/alpine:3.20 \
  -t sms-server-go .
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
- `scripts/deploy_remote_podman.ps1`（SSH 直连远端 Podman 构建并部署，推荐）
- `deploy/docker/deploy.sh`（服务器执行部署，含健康检查与失败回滚）
- `deploy/docker/rollback.sh`（手动回滚）

脚本流程：

1. `go test ./...`
2. `docker build` 构建镜像（建议使用镜像源参数，避免走 Docker 官方源）
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

> 说明：该脚本会在本地先执行 `docker build`。若本地无法访问 Docker 官方源，建议优先使用下方“SSH 直连部署（Podman）”方式，或先确保本地 Docker 已配置镜像加速。

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\deploy_remote.ps1 `
  -RemoteHost 203.0.113.10 `
  -RemoteUser root `
  -RemotePath /opt/sms-server `
  -ImageTag v1.0.0
```

可选参数：

- `-IdentityFile C:\Users\you\.ssh\id_rsa_new`
- `-SshPort 22`
- `-HealthUrl http://127.0.0.1:8322/sms/healthz`
- `-HealthTimeout 90`
- `-SkipTests`

### SSH 直连部署（Podman，推荐）

推荐直接使用脚本：

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\deploy_remote_podman.ps1 `
  -RemoteHost 120.48.167.92 `
  -RemoteUser root `
  -RemotePath /opt/sms-server `
  -IdentityFile "$HOME\.ssh\id_rsa_new"
```

该脚本会自动执行：

1. 本地打包源码（排除 `.git`、`build`、`data`、`node_modules` 等）
2. `scp` 上传到远端目录
3. 远端 `podman build`（默认使用 daocloud 基础镜像）
4. 远端重建并启动容器
5. 远端健康检查（默认 `http://127.0.0.1:8322/sms/healthz`）

常用可选参数：

- `-SshPort 22`
- `-ImageRef localhost/sms-server-go:latest`
- `-ContainerName sms-server`
- `-HealthUrl http://127.0.0.1:8322/sms/healthz`
- `-HealthTimeout 120`
- `-BindAddress 127.0.0.1`
- `-HostPort 8322`
- `-GoBaseImage docker.m.daocloud.io/library/golang:1.22-alpine`
- `-RuntimeBaseImage docker.m.daocloud.io/library/alpine:3.20`
- `-NodeBaseImage docker.m.daocloud.io/library/node:20-alpine`

### SSH 直连部署（Podman，手工方式）

当你不想使用脚本时，可手工执行：

```powershell
scp -i "$HOME\.ssh\id_rsa_new" -r .\* root@120.48.167.92:/opt/sms-server/
ssh -i "$HOME\.ssh\id_rsa_new" root@120.48.167.92
```

```bash
cd /opt/sms-server
cp -n .env.example .env
mkdir -p data

podman build \
  --build-arg GO_BASE_IMAGE=docker.m.daocloud.io/library/golang:1.22-alpine \
  --build-arg RUNTIME_BASE_IMAGE=docker.m.daocloud.io/library/alpine:3.20 \
  --build-arg NODE_BASE_IMAGE=docker.m.daocloud.io/library/node:20-alpine \
  -t localhost/sms-server-go:latest .

podman rm -f sms-server || true
podman run -d --name sms-server \
  --restart unless-stopped \
  --env-file .env \
  -p 127.0.0.1:8322:8322 \
  -v /opt/sms-server/data:/app/data \
  localhost/sms-server-go:latest

curl -fsS http://127.0.0.1:8322/sms/healthz
```

说明：

- 该流程不依赖本地 Docker，只需要本机可用 `ssh/scp`。
- 为避免走 Docker 官方源，建议保留上面的 `daocloud` build-arg。
- `frontend/static/app` 是 `npm run build:ui` 的构建产物，**不入库**（见 `.gitignore`）；
  镜像内由 Dockerfile 的 Node 阶段自动生成，因此只要改动 `src/` 就会在重建时生效。
- 改动了 `frontend/static`（非 `app` 子目录，如 `js/`、`css/`）后，也必须重建镜像并重启容器。

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

    location ^~ /sms/ {
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

    location ^~ /v1/sms/ {
        proxy_pass http://127.0.0.1:8322/v1/sms/;
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

若站点还有静态资源正则（例如 `location ~* \.(js|css|png...)`），请保留 `^~ /sms/` 前缀，避免 `/sms/assets/*.js` 被静态规则抢占导致 404。

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

近期 UI 改造：

- 前端架构升级为 React + TypeScript + Rsbuild
- 拆分功能模块与设置模块，提升操作清晰度
- 增强移动端适配，下载按钮收敛到标题栏（移动端直接下载，桌面端弹窗显示二维码）

构建加速（Dockerfile 分层缓存优化）：

- 前端阶段将 `npm ci` 与 `npm run build:ui` 拆成两层，仅当 `package*.json` 变化时才重装依赖。
- Go 阶段只复制 `cmd/`、`internal/`、`go.mod`、`go.sum`、`vendor/`，前端改动不再让 Go 编译层缓存失效。
- 对 `npm ci` 与 `go build` 启用 `--mount=type=cache` 缓存挂载，加速依赖变化时的增量构建。
- 效果：全缓存命中的重复部署从约 180s 降至约 20s；仅改前端时跳过 Go 编译，仅改后端时跳过前端依赖安装。

## 许可证

MIT
