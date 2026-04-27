package main

import (
	"log"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"cinema-reservation/gateway/internal"
	pb "cinema-reservation/common/proto/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Connessione al Catalog Service
	catAddr := os.Getenv("CATALOG_SERVICE_ADDR")
	if catAddr == "" { catAddr = "localhost:50052" }
	connCat, _ := grpc.Dial(catAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))

	// Connessione alle istanze di Booking Service (per lo sharding)
	// In produzione qui leggeresti da un service discovery
	bookAddrs := strings.Split(os.Getenv("BOOKING_SERVICE_ADDRS"), ",")
	if len(bookAddrs) == 1 && bookAddrs[0] == "" { bookAddrs = []string{"localhost:50051"} }

	var bookClients []pb.BookingServiceClient
	for _, addr := range bookAddrs {
		conn, _ := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		bookClients = append(bookClients, pb.NewBookingServiceClient(conn))
	}

	handler := internal.NewGatewayHandler(pb.NewCatalogServiceClient(connCat), bookClients)

	r := gin.Default()
	r.GET("/movies", handler.GetMovies)
	r.POST("/book", handler.ReserveSeat)

	log.Println("Gateway in ascolto su :8080")
	r.Run(":8080")
}
