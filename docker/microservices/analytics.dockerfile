FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o analytics-service analytics.go


# -----------------------------
FROM alpine:latest

WORKDIR /app

RUN adduser -D analyticsuser
USER analyticsuser

COPY --from=builder /app/analytics-service .

EXPOSE 4800

CMD ["./analytics-service"]
