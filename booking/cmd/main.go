package main

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"

	"cinema-reservation/booking/internal"
	"cinema-reservation/common/utils"
	pb "cinema-reservation/common/proto/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// =========================
	// CONFIG
	// =========================
	port := utils.GetEnv("PORT", "50051")
	natsURL := utils.GetEnv("NATS_URL", nats.DefaultURL)
	redisAddr := utils.GetEnv("REDIS_ADDR", "redis:6379")
	catalogPath := utils.GetEnv("CATALOG_FILE", "/app/movies.json")

	// =========================
	// NATS
	// =========================
	nc, err := nats.Connect(
		natsURL,
		nats.Name("booking-service"),
		nats.MaxReconnects(-1),
		nats.ReconnectWait(2*time.Second),
	)
	if err != nil {
		log.Fatalf("NATS connection error: %v", err)
	}
	defer nc.Drain()

	// =========================
	// REDIS
	// =========================
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	timeout := utils.GetEnvDuration("REDIS_TIMEOUT", 2*time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Redis connection error: %v", err)
	}

	log.Println(" Connected to Redis:", redisAddr)

	// =========================
	// gRPC
	// =========================
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Listen error: %v", err)
	}

	s := grpc.NewServer()

	bookingService := internal.NewBookingService(nc, rdb, catalogPath)

	pb.RegisterBookingServiceServer(s, bookingService)

	reflection.Register(s)

	log.Printf(" Booking Service running on :%s", port)
	log.Printf(" NATS: %s", natsURL)
	log.Printf(" Redis: %s", redisAddr)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
