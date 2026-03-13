FROM golang:1.22-alpine AS builder

ARG ALPINE_REPO_HOST=mirrors.aliyun.com
RUN sed -i "s/dl-cdn.alpinelinux.org/${ALPINE_REPO_HOST}/g" /etc/apk/repositories

WORKDIR /src

RUN apk add --no-cache git

COPY go.mod go.sum ./
ARG GO_PROXY=https://goproxy.cn,direct
ENV GOPROXY=${GO_PROXY}
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/server ./cmd/server

FROM alpine:3.20

ARG ALPINE_REPO_HOST=mirrors.aliyun.com
RUN sed -i "s/dl-cdn.alpinelinux.org/${ALPINE_REPO_HOST}/g" /etc/apk/repositories

RUN apk add --no-cache ca-certificates tzdata su-exec && adduser -D -H -u 10001 appuser

WORKDIR /app

COPY --from=builder /out/server /app/server
COPY docker-entrypoint.sh /app/docker-entrypoint.sh

RUN mkdir -p /app/data/uploads && chown -R appuser:appuser /app && chmod +x /app/docker-entrypoint.sh

EXPOSE 8080

ENTRYPOINT ["/app/docker-entrypoint.sh"]
