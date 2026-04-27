FROM golang:1.25.9-alpine AS builder
RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Compila il servizio Booking
RUN CGO_ENABLED=0 GOOS=linux go build -o /booking-service ./booking/cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/
COPY --from=builder /booking-service .

# Porta gRPC
EXPOSE 50051

CMD ["./booking-service"]
