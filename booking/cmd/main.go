package main

import (
	"log"
	"net"
	"os"

	"github.com/nats-io/nats.go"
	"cinema-reservation/booking/internal"
	pb "cinema-reservation/common/proto/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// 1. Connessione a NATS
	// Usiamo una variabile d'ambiente per l'URL (fondamentale per Docker)
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = nats.DefaultURL // "nats://127.0.0.1:4222"
	}

	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatalf("Impossibile connettersi a NATS: %v", err)
	}
	defer nc.Close()

	// 2. Setup gRPC
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	
	// Passiamo la connessione NATS al servizio
	pb.RegisterBookingServiceServer(s, internal.NewBookingService(nc,"movies.json"))
	
	reflection.Register(s)

	log.Printf("Booking Service in ascolto su :50051 con NATS su %s", natsURL)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
