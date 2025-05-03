    FROM golang:1.22-alpine AS builder

    WORKDIR /app
    
    COPY ../../micro-services/shoppingcart-service/go.mod ../../micro-services/shoppingcart-service/go.sum ./
    
    RUN go mod download
    
    COPY ../../micro-services/shoppingcart-service/ ./
    
    RUN go build -o shoppingcart-service cart.go
    
   
    # -----------------------------
    FROM alpine:latest
    
    WORKDIR /app

    RUN adduser -D cartuser
    USER cartuser
  
    COPY --from=builder /app/shoppingcart-service .
    
    EXPOSE 4300
    
    CMD ["./shoppingcart-service"]
    