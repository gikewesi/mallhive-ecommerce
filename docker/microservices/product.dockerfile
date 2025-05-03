FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY ../../micro-services/product-service/go.mod ../../micro-services/product-service/go.sum ./

RUN go mod download

COPY ../../micro-services/product-service/ ./

RUN go build -o product-service product.go


# -----------------------------
FROM alpine:latest

WORKDIR /app

RUN adduser -D productuser
USER productuser

COPY --from=builder /app/product-service .

EXPOSE 4100

CMD ["./product-service"]
