ARG GO_BASE_IMAGE=docker.io/library/golang:1.22-alpine
ARG RUNTIME_BASE_IMAGE=docker.io/library/alpine:3.20
ARG GO_PROXY=https://goproxy.cn,direct
ARG GO_SUMDB=sum.golang.google.cn
ARG ALPINE_MIRROR=https://mirrors.aliyun.com/alpine

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

COPY . .
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -mod=vendor -trimpath -ldflags="-s -w" -o /app/sms-server ./cmd/server

FROM ${RUNTIME_BASE_IMAGE}
ARG ALPINE_MIRROR
RUN sed -i "s#https://dl-cdn.alpinelinux.org/alpine#${ALPINE_MIRROR}#g" /etc/apk/repositories \
    && apk add --no-cache ca-certificates tzdata
WORKDIR /app

COPY --from=builder /app/sms-server /app/sms-server
COPY --from=builder /app/frontend/templates/ui.html /app/frontend/templates/ui.html
COPY --from=builder /app/frontend/static /app/frontend/static
RUN mkdir -p /app/data

EXPOSE 8322

CMD ["/app/sms-server"]
