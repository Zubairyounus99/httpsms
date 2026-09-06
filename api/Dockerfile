# Stage 1: Build the Go binary
FROM golang:latest AS builder

WORKDIR /http-sms

# Pre-copy and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Install Swagger CLI
RUN go get github.com/swaggo/swag/gen@latest && \
    go get github.com/swaggo/swag/cmd/swag@latest && \
    go install github.com/swaggo/swag/cmd/swag

# Generate swagger documentation
RUN swag init --requiredByDefault --parseDependency --parseInternal

# Patch Axiom crash directly in container.go to fall back to stdout console logger
RUN sed -i 's/return zerolog.Logger{}, errors.Wrap(err, "cannot create axiom zerolog writer")/output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}; return zerolog.New(output).With().Timestamp().Logger(), nil/g' pkg/di/container.go

# Compile binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-X main.Version=$GIT_COMMIT" -o /bin/http-sms .

# Stage 2: Runtime image
FROM alpine:latest

RUN apk add --no-cache curl ca-certificates tzdata && \
    addgroup -S http-sms && adduser -S http-sms -G http-sms

WORKDIR /home/http-sms

COPY --from=builder /usr/local/go/lib/time/zoneinfo.zip /zoneinfo.zip
COPY --from=builder /bin/http-sms ./
COPY --from=builder /http-sms/root.crt ./

USER http-sms

EXPOSE 8000

CMD ["./http-sms"]
