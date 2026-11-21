FROM golang:1.26-alpine AS builder

WORKDIR /src
RUN apk add --no-cache ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/chatapp ./cmd

FROM builder AS migrator-build
RUN CGO_ENABLED=0 GOBIN=/out go install github.com/pressly/goose/v3/cmd/goose@v3.24.1

FROM alpine:3.22 AS migrator
RUN apk add --no-cache ca-certificates
COPY --from=migrator-build /out/goose /usr/local/bin/goose
COPY --from=builder /src/internal/database/migrations /migrations
ENTRYPOINT ["/bin/sh", "-ec"]

FROM alpine:3.22 AS runtime
RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S app \
    && adduser -S -G app app
WORKDIR /app
COPY --from=builder /out/chatapp /app/chatapp
USER app
EXPOSE 8080
ENTRYPOINT ["/app/chatapp"]
