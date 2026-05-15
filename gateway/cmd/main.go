package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"cinema-reservation/common/utils"
	pb "cinema-reservation/common/proto/pb"
	"cinema-reservation/gateway/internal"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	_ "google.golang.org/grpc/balancer/roundrobin"
)

func main() {

        //connection to Catalog service
	catAddr := utils.GetEnv("CATALOG_SERVICE_ADDR", "catalog-service:50052")

	connCat, err := grpc.Dial(
		"dns:///"+catAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
	)
	if err != nil {
		log.Fatalf("Catalog connection error: %v", err)
	}

	//connection to booking service
	bookAddr := utils.GetEnv("BOOKING_SERVICE_ADDR", "booking:50051")

	connBook, err := grpc.Dial(
		"dns:///"+bookAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
	)
	if err != nil {
		log.Fatalf("Booking connection error: %v", err)
	}
         
        //connection to recommendation service
	recAddr := utils.GetEnv("RECOMMENDATION_SERVICE_ADDR", "recommendation-service:50053")

	connRec, err := grpc.Dial(
		"dns:///"+recAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
	)
	if err != nil {
		log.Fatalf("Recommendation connection error: %v", err)
	}

	handler := internal.NewGatewayHandler(
		pb.NewCatalogServiceClient(connCat),
		pb.NewBookingServiceClient(connBook),
		pb.NewRecommendationServiceClient(connRec),
	)


	r := gin.Default()

	r.GET("/movies", handler.GetMovies)
	r.GET("/seats/:id", handler.GetSeats)
	r.POST("/book", handler.ReserveSeat)

	r.GET("/recommendations", handler.GetRecommendations)

	
	r.GET("/projections", handler.GetProjections)
	r.GET("/projections/:id", handler.GetProjectionsByMovie)

	log.Println("API Gateway running on :8080 with gRPC load balancing enabled")
	r.Run(":8080")
}
