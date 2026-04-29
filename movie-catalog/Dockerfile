FROM golang:1.25.9-alpine AS builder
RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /catalog-service ./movie-catalog/cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# 1. Copia il binario compilato dallo stage builder
COPY --from=builder /catalog-service .

# 2. COPIA QUI IL FILE JSON 
# Assicurati che il percorso sorgente sia corretto rispetto alla root del build context
COPY ./movies.json .

# Porta gRPC
EXPOSE 50052

CMD ["./catalog-service"]
