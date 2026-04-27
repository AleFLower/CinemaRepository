package main

import (
	"log"
	"net"

	"cinema-reservation/movie-catalog/internal"
	pb "cinema-reservation/common/proto/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("Errore apertura porta 50052: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterCatalogServiceServer(s, internal.NewCatalogService())
	
	// Abilitiamo la reflection per i test con grpcurl
	reflection.Register(s)

	log.Printf("Movie Catalog Service in ascolto su :50052")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Errore avvio server: %v", err)
	}
}
