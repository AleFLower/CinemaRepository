# Stage 1: Build
FROM golang:1.25.9-alpine AS builder

# Installa git per eventuali dipendenze
RUN apk add --no-cache git

WORKDIR /app

# Copia i file dei moduli dalla root (necessario perché il progetto è un modulo unico)
COPY go.mod go.sum ./
RUN go mod download

# Copia tutto il codice del progetto
COPY . .

# Compila il gateway
RUN CGO_ENABLED=0 GOOS=linux go build -o /gateway ./gateway/cmd/main.go

# Stage 2: Final image
FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/
COPY --from=builder /gateway .

# Espone la porta del Gateway (es. 8080)
EXPOSE 8080

CMD ["./gateway"]
