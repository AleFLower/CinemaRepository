package main

import (
	"log"
	"net"

	"cinema-reservation/common/utils"
	"cinema-reservation/movie-catalog/internal"
	pb "cinema-reservation/common/proto/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// Porta configurabile (fallback default)
	port := utils.GetEnv("PORT", "50052")

	// Path file configurabile
	catalogPath := utils.GetEnv("CATALOG_FILE", "movies.json")

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Errore apertura porta %s: %v", port, err)
	}

	s := grpc.NewServer()

	pb.RegisterCatalogServiceServer(s, internal.NewCatalogService(catalogPath))

	reflection.Register(s)

	log.Printf("Movie Catalog Service in ascolto su :%s", port)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Errore avvio server: %v", err)
	}
}
