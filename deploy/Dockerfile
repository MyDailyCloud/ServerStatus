# 多阶段构建
FROM golang:1.25.5-alpine AS builder

ARG VERSION=dev
ARG GIT_COMMIT=unknown
ARG BUILD_TIME=unknown

WORKDIR /app

# 预下载根模块依赖（monitor-agent 使用）
COPY go.mod go.sum ./
RUN go mod download

# 预下载 data-server 模块依赖
COPY data-server/go.mod data-server/go.sum ./data-server/
RUN cd data-server && go mod download

# 复制全部源码
COPY . .

# 构建 data-server（使用 data-server 模块）
RUN cd data-server && CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-s -w -X 'main.buildVersion=${VERSION}' -X 'main.buildCommit=${GIT_COMMIT}' -X 'main.buildTime=${BUILD_TIME}'" \
    -o /app/bin/data-server .

# 构建 monitor-agent（使用根模块）
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-s -w -X 'main.buildVersion=${VERSION}' -X 'main.buildCommit=${GIT_COMMIT}' -X 'main.buildTime=${BUILD_TIME}'" \
    -o /app/bin/monitor-agent ./monitor-agent

# 最终镜像
FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY --from=builder /app/bin/data-server .
COPY --from=builder /app/bin/monitor-agent .
COPY --from=builder /app/data-server/web-ui ./web-ui
COPY --from=builder /app/data-server/server-config.json .
COPY --from=builder /app/monitor-agent/config.json ./agent-config.json

EXPOSE 8080

CMD ["./data-server"]