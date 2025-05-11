FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o product-service product.go

# -----------------------------
FROM alpine:latest

WORKDIR /app

RUN adduser -D productuser
USER productuser

COPY --from=builder /app/product-service .

EXPOSE 4100

CMD ["./product-service"]
