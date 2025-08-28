FROM golang:1.24-alpine AS builder

# goose
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

FROM alpine:latest

# postgresql-client
RUN apk add --no-cache postgresql-client

# Copy goose binary from builder
COPY --from=builder /go/bin/goose /usr/local/bin/goose

COPY migrations/ /migrations/

COPY migrations/migrate.sh /migrate.sh
RUN chmod +x /migrate.sh

ENTRYPOINT ["/migrate.sh"]
