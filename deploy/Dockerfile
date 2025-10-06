# 多阶段构建
FROM golang:1.24.3-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w" -o data-server ./data-server/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w" -o monitor-agent ./monitor-agent/main.go

# 最终镜像
FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY --from=builder /app/data-server .
COPY --from=builder /app/monitor-agent .
COPY --from=builder /app/data-server/web-ui ./web-ui
COPY --from=builder /app/data-server/server-config.json .
COPY --from=builder /app/monitor-agent/config.json ./agent-config.json

EXPOSE 8080

CMD ["./data-server"]