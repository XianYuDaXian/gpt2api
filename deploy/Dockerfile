# syntax=docker/dockerfile:1.6

FROM golang:1.26-alpine AS backend-builder

WORKDIR /src
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o /out/gpt2api ./cmd/server
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    GOBIN=/out CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go install -ldflags "-s -w" github.com/pressly/goose/v3/cmd/goose@v3.20.0

FROM node:22-alpine AS web-builder

WORKDIR /src/web
COPY web/package*.json ./
RUN --mount=type=cache,target=/root/.npm npm ci --no-audit --no-fund
COPY web/ ./
RUN npm run build

FROM alpine:3.20

# 走 ustc HTTP 源,规避企业 HTTPS 代理 MITM 导致的证书校验失败
RUN sed -i 's@https\?://dl-cdn.alpinelinux.org@http://mirrors.ustc.edu.cn@g' /etc/apk/repositories \
    && apk add --no-cache ca-certificates tzdata curl bash mariadb-client \
    && update-ca-certificates \
    && ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone

WORKDIR /app

# 后端 + goose 二进制
COPY --from=backend-builder /out/gpt2api /app/gpt2api
COPY --from=backend-builder /out/goose   /usr/local/bin/goose
RUN chmod +x /app/gpt2api /usr/local/bin/goose

# 前端构建产物 —— 被 Go server 以 SPA 方式托管
COPY --from=web-builder /src/web/dist /app/web/dist

# Migrations、默认配置、entrypoint
COPY sql     /app/sql
COPY configs /app/configs
COPY deploy/entrypoint.sh /app/entrypoint.sh
RUN sed -i 's/\r$//' /app/entrypoint.sh \
    && chmod +x /app/entrypoint.sh \
    && mkdir -p /app/data/backups /app/logs

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=20s --retries=3 \
    CMD curl -fsS http://localhost:8080/healthz || exit 1

ENTRYPOINT ["/app/entrypoint.sh"]
CMD ["/app/gpt2api", "-c", "/app/configs/config.yaml"]
