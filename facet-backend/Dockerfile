# --- Build Stage ---
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

ARG GOPROXY="https://proxy.golang.org,direct"
ENV GOPROXY=${GOPROXY}

RUN go mod download

COPY . .

ENV GOMAXPROCS=1
ENV CGO_ENABLED=0
ENV GOOS=linux
RUN go build -p 1 -ldflags="-w -s" -o /server ./cmd/server/main.go

# --- Final Stage ---
FROM alpine:3.20.1

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /server .
COPY config.yaml .

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

RUN chown -R appuser:appgroup /app

USER appuser

EXPOSE 8080

CMD ["./server"]
