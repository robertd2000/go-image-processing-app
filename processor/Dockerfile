# =========================
# Stage 1: Build
# =========================
FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# build
RUN go build -o app cmd/main.go

# =========================
# Stage 2: Runtime
# =========================
FROM alpine:3.19

WORKDIR /app

# timezone + certs (важно для JWT / HTTPS)
RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/app .

COPY --from=builder /app/config ./config

COPY --from=builder /app/docs ./docs

# port
EXPOSE 8083

# run
CMD ["./app"]
