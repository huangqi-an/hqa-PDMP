FROM golang:1.27-alpine AS builder

ENV GOPROXY=https://goproxy.cn,direct \
  CGO_ENABLED=0 \
  GOOS=linux

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal

RUN go build -o /out/worker-service ./cmd/server

FROM alpine:3.22

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /out/worker-service ./worker-service

EXPOSE 8082

CMD ["./worker-service"]