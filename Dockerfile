FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git nodejs npm

WORKDIR /build

COPY go.mod go.sum ./apps/server/
WORKDIR /build/apps/server
RUN go mod download

COPY . /build

WORKDIR /build/apps/web
RUN npm install -g pnpm && pnpm install && pnpm build

WORKDIR /build/apps/server
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /wui-server cmd/main.go

FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /wui-server /app/
COPY --from=builder /build/apps/web/dist /app/web

RUN mkdir -p /app/data /app/xray /app/logs

EXPOSE 32452

ENV WUI_CONFIG=/app/data/config.json

CMD ["/app/wui-server"]
