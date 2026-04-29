package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"cinema-reservation/gateway/internal"
	pb "cinema-reservation/common/proto/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {

	// 📚 Catalog Service
	catAddr := os.Getenv("CATALOG_SERVICE_ADDR")
	if catAddr == "" {
		catAddr = "catalog-service:50052"
	}

	connCat, err := grpc.Dial(
	catAddr,
	grpc.WithTransportCredentials(insecure.NewCredentials()),
	grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
)
	if err != nil {
		log.Fatalf("Errore connessione catalog: %v", err)
	}

	// 🎟️ Booking Service (scalabile)
	bookAddr := os.Getenv("BOOKING_SERVICE_ADDR")
	if bookAddr == "" {
		bookAddr = "booking:50051"
	}

	connBook, err := grpc.Dial(bookAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Errore connessione booking: %v", err)
	}

	handler := internal.NewGatewayHandler(
		pb.NewCatalogServiceClient(connCat),
		pb.NewBookingServiceClient(connBook),
	)

	// 🌐 HTTP Server
	r := gin.Default()

	r.GET("/movies", handler.GetMovies)
	r.GET("/seats/:id", handler.GetSeats)
	r.POST("/book", handler.ReserveSeat)

	log.Println("🚀 API Gateway in ascolto su :8080")
	r.Run(":8080")
}
