FROM golang:1.25.4-alpine AS builder

RUN apk add --no-cache git 

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o app ./cmd/app/main.go
FROM alpine:latest

RUN apk --no-cache add ca-certificates postgresql-client

WORKDIR /app

COPY --from=builder /build/app /app/app
COPY --from=builder /go/bin/migrate /usr/local/bin/migrate
COPY --from=builder /build/scripts/migrate.sh /app/scripts/migrate.sh
# COPY --from=builder /build/.env /app/.env
COPY --from=builder /build/migrations /app/migrations

RUN chmod +x /app/scripts/migrate.sh

CMD ["/app/app"]

