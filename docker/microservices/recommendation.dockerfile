FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY ../../micro-services/recommendation-service/go.mod ../../micro-services/recommendation-service/go.sum ./

RUN go mod download

COPY ../../micro-services/recommendation-service/ ./

RUN go build -o recommendation-service recommendation.go


# -----------------------------
FROM alpine:latest

WORKDIR /app

RUN adduser -D recommendationuser
USER recommendationuser

COPY --from=builder /app/recommendation-service .

EXPOSE 4700

CMD ["./recommendation-service"]
