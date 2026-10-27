ARG GO_BASE_IMAGE=docker.io/library/golang:1.22-alpine
ARG RUNTIME_BASE_IMAGE=docker.io/library/alpine:3.20
ARG NODE_BASE_IMAGE=docker.io/library/node:20-alpine
ARG GO_PROXY=https://goproxy.cn,direct
ARG GO_SUMDB=sum.golang.google.cn
ARG ALPINE_MIRROR=https://mirrors.aliyun.com/alpine

FROM ${NODE_BASE_IMAGE} AS frontend-builder
WORKDIR /app

# 依赖层：仅当 package*.json 变化时才重跑 npm ci（配合 npm 缓存挂载加速）
COPY package.json package-lock.json ./
RUN --mount=type=cache,target=/root/.npm npm ci

# 源码层：仅前端源码变化时才重新构建（npm ci 保持缓存命中）
COPY tsconfig.json rsbuild.config.ts ./
COPY src ./src
RUN npm run build:ui

FROM ${GO_BASE_IMAGE} AS builder
ARG GO_PROXY
ARG GO_SUMDB
ARG ALPINE_MIRROR
WORKDIR /app
ENV GOPROXY=${GO_PROXY}
ENV GOSUMDB=${GO_SUMDB}
RUN sed -i "s#https://dl-cdn.alpinelinux.org/alpine#${ALPINE_MIRROR}#g" /etc/apk/repositories \
    && apk add --no-cache build-base

COPY go.mod go.sum ./
COPY vendor ./vendor

# 仅复制 Go 源码，前端改动不会让 Go 编译层缓存失效
COPY cmd ./cmd
COPY internal ./internal
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -mod=vendor -trimpath -ldflags="-s -w" -o /app/sms-server ./cmd/server

# 前端静态资源在 Go 编译之后复制（纯拷贝，不触发重新编译）
COPY frontend/static ./frontend/static
COPY --from=frontend-builder /app/frontend/static/app ./frontend/static/app

FROM ${RUNTIME_BASE_IMAGE}
ARG ALPINE_MIRROR
RUN sed -i "s#https://dl-cdn.alpinelinux.org/alpine#${ALPINE_MIRROR}#g" /etc/apk/repositories \
    && apk add --no-cache ca-certificates tzdata
WORKDIR /app

COPY --from=builder /app/sms-server /app/sms-server
COPY --from=builder /app/frontend/static /app/frontend/static
RUN mkdir -p /app/data

EXPOSE 8322

CMD ["/app/sms-server"]
