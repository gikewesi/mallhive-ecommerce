FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o shoppingcart-service cart.go

# -----------------------------
FROM alpine:latest

WORKDIR /app

RUN adduser -D cartuser
USER cartuser

COPY --from=builder /app/shoppingcart-service .

EXPOSE 4300

CMD ["./shoppingcart-service"]
