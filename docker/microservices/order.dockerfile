FROM golang:1.23 AS builder

WORKDIR /app

COPY ./go.mod ./go.sum ./

RUN go mod download

COPY . ./

RUN go build -o order-service order.go


# -----------------------------
FROM alpine:latest

WORKDIR /app

RUN adduser -D orderuser
USER orderuser

COPY --from=builder /app/order-service .

EXPOSE 4200

CMD ["./order-service"]
