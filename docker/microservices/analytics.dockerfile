FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY ../../micro-services/analystics-service/go.mod ../../micro-services/analystics-service/go.sum ./

RUN go mod download

COPY ../../micro-services/analystics-service/ ./

RUN go build -o analystics-service analystics.go


# -----------------------------
FROM alpine:latest

WORKDIR /app

RUN adduser -D analysticsuser
USER analysticsuser

COPY --from=builder /app/analystics-service .

EXPOSE 4800

CMD ["./analystics-service"]
