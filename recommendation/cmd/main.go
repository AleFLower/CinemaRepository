package main

import (
	"context"
	"log"
	"net"

	"cinema-reservation/common/utils"
	pb "cinema-reservation/common/proto/pb"
	"cinema-reservation/recommendation/internal"

	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {

	
	port := utils.GetEnv("PORT", "50053")
	natsURL := utils.GetEnv("NATS_URL", "nats://localhost:4222")
	redisAddr := utils.GetEnv("REDIS_ADDR", "localhost:6379")
	catalogAddr := utils.GetEnv("CATALOG_SERVICE_ADDR", "localhost:50052")


	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatalf("NATS connection error: %v", err)
	}
	defer nc.Close()

	log.Println("[NATS] connected")

	
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	ctx := context.Background()

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Redis connection error: %v", err)
	}

	log.Println("[REDIS] connected")


	connCat, err := grpc.Dial(
		catalogAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("Catalog connection error: %v", err)
	}
	defer connCat.Close()

	catalogClient := pb.NewCatalogServiceClient(connCat)

	log.Println("[CATALOG] connected")


	service := internal.NewRecommendationService(
		nc,
		rdb,
		catalogClient,
	)


	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Listen error: %v", err)
	}

	grpcServer := grpc.NewServer()

	pb.RegisterRecommendationServiceServer(grpcServer, service)

	log.Printf("[RECOMMENDATION] running on port %s", port)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
