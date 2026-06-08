# syntax=docker/dockerfile:1
# Baize Wiki — Multi-stage build
# Stage 1: Build the Go binary
FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /baize-wiki ./cmd/baize-wiki

# Stage 2: Minimal runtime image
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /baize-wiki /usr/local/bin/baize-wiki

EXPOSE 9876
ENTRYPOINT ["baize-wiki"]
CMD ["--help"]
