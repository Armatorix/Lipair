FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o bin/server ./cmd/server

FROM alpine:3.20

# golang:1.22-alpine includes ca-certificates at /etc/ssl/certs/ca-certificates.crt.
# We copy it here to avoid running apk (which requires outbound network access).
# Timezone data is embedded directly in the binary via `import _ "time/tzdata"`.
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

WORKDIR /app

COPY --from=builder /app/bin/server .

EXPOSE 8080

CMD ["./server"]
