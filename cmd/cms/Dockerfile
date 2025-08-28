FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

RUN go install github.com/pressly/goose/v3/cmd/goose@latest

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/cms/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY --from=builder /app/main .

COPY --from=builder /go/bin/goose /usr/local/bin/goose
COPY migrations/ migrations/

EXPOSE 3000

HEALTHCHECK --interval=5s --timeout=3s CMD curl --fail http://localhost:3000/healthz || exit 1
CMD ["./main"]

