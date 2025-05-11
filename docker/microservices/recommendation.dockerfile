FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o recommendation-service recommendation.go

# -----------------------------
FROM alpine:latest

WORKDIR /app

RUN adduser -D recommendationuser
USER recommendationuser

COPY --from=builder /app/recommendation-service .

EXPOSE 4700

CMD ["./recommendation-service"]
