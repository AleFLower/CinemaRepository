FROM golang:1.25.9-alpine AS builder
RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /catalog-service ./movie-catalog/cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /catalog-service .

COPY movies.json /app/movies.json

EXPOSE 50052

CMD ["./catalog-service"]
