package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"cinema-reservation/common/utils"
	"cinema-reservation/gateway/internal"
	pb "cinema-reservation/common/proto/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	_ "google.golang.org/grpc/balancer/roundrobin"
)

func main() {

	// Catalog Service
	catAddr := utils.GetEnv("CATALOG_SERVICE_ADDR", "catalog-service:50052")

	connCat, err := grpc.Dial(
		"dns:///"+catAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
	)
	if err != nil {
		log.Fatalf("Errore connessione catalog: %v", err)
	}

	// Booking Service
	bookAddr := utils.GetEnv("BOOKING_SERVICE_ADDR", "booking:50051")

	connBook, err := grpc.Dial(
		"dns:///"+bookAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
	)
	if err != nil {
		log.Fatalf("Errore connessione booking: %v", err)
	}

	handler := internal.NewGatewayHandler(
		pb.NewCatalogServiceClient(connCat),
		pb.NewBookingServiceClient(connBook),
	)

	r := gin.Default()
	r.GET("/movies", handler.GetMovies)
	r.GET("/seats/:id", handler.GetSeats)
	r.POST("/book", handler.ReserveSeat)

	log.Println("API Gateway in ascolto su :8080 con gRPC Load Balancing")
	r.Run(":8080")
}
