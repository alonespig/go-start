# ---- 构建阶段 ----
FROM golang:1.26-alpine AS builder

WORKDIR /app

# 先复制依赖描述文件，利用 Docker 层缓存
COPY go.mod go.sum ./
RUN go mod download

# 复制源码并编译
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/server .

# ---- 运行阶段 ----
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S app && adduser -S app -G app

WORKDIR /app

COPY --from=builder /app/server .
COPY config.yaml .

USER app

EXPOSE 8080

ENTRYPOINT ["./server"]
CMD ["serve"]
