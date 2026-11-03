# syntax=docker/dockerfile:1
FROM golang:1.26-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /app/server ./cmd

FROM alpine:3.23 AS runtime

RUN apk --no-cache add ca-certificates tzdata \
    && addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app
COPY --from=builder /app/server /app/server

RUN touch .env && mkdir -p config/logs && chown -R appuser:appgroup /app

USER appuser
EXPOSE 8888
ENTRYPOINT ["/app/server"]
