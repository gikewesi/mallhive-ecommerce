FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY ../../micro-services/order-service/go.mod ../../micro-services/order-service/go.sum ./

RUN go mod download

COPY ../../micro-services/order-service/ ./

RUN go build -o order-service order.go


# -----------------------------
FROM alpine:latest

WORKDIR /app

RUN adduser -D orderuser
USER orderuser

COPY --from=builder /app/order-service .

EXPOSE 4200

CMD ["./order-service"]
