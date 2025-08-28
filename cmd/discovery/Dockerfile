FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o discovery cmd/discovery/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY --from=builder /app/discovery .

EXPOSE 4000

HEALTHCHECK --interval=5s --timeout=3s CMD curl --fail http://localhost:4000/healthz || exit 1

CMD ["./discovery"]

