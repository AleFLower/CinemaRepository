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
	// Configurable port (default fallback)
	port := utils.GetEnv("PORT", "50052")

	// Configurable file path
	catalogPath := utils.GetEnv("CATALOG_FILE", "movies.json")

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Error opening port %s: %v", port, err)
	}

	s := grpc.NewServer()

	pb.RegisterCatalogServiceServer(s, internal.NewCatalogService(catalogPath))

	reflection.Register(s)

	log.Printf("Movie Catalog Service listening on :%s", port)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Server startup error: %v", err)
	}
}
