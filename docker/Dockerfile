FROM golang:1.25-alpine AS builder

WORKDIR /app

ENV GOPROXY=https://goproxy.cn,direct

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o ./admin-seed cmd/admin-seed/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o ./homelab-blog main.go

FROM alpine:3.24

ARG UID=1000
ARG GID=1000

RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -g $GID -S appgroup && \
    adduser -u $UID -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /app/admin-seed .

COPY --from=builder /app/homelab-blog .

RUN mkdir -p /app/data/uploads && chown -R appuser:appgroup /app

USER appuser

EXPOSE 5200

ENTRYPOINT ["./homelab-blog"]