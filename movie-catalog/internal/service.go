package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	pb "cinema-reservation/common/proto/pb"
)

type CatalogService struct {
	pb.UnimplementedCatalogServiceServer
	movies []pb.Movie
}

func NewCatalogService(configPath string) *CatalogService {
	// 1. Leggiamo il file JSON
	file, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalf("Errore nel caricamento del file catalog: %v", err)
	}

	// 2. Parsiamo il JSON nello slice di movies
	var movies []pb.Movie
	err = json.Unmarshal(file, &movies)
	if err != nil {
		log.Fatalf("Errore nel parsing del JSON: %v", err)
	}

	log.Printf("Catalog caricato: %d film trovati", len(movies))

	return &CatalogService{
		movies: movies,
	}
}

func (s *CatalogService) GetMovies(ctx context.Context, req *pb.Empty) (*pb.MovieList, error) {
log.Println("Catalog instance serving request")
	log.Println("Catalog instance serving request")
	
	// Trasformiamo lo slice di struct in un puntatore a MovieList
	// Usiamo i puntatori perché il file .pb.go generato si aspetta []*Movie
	moviePtrs := make([]*pb.Movie, len(s.movies))
	for i := range s.movies {
		moviePtrs[i] = &s.movies[i]
	}
	return &pb.MovieList{Movies: moviePtrs}, nil
}

func (s *CatalogService) GetMovie(ctx context.Context, req *pb.MovieRequest) (*pb.Movie, error) {
	for _, m := range s.movies {
		if m.Id == req.Id {
			// Attenzione: m è una copia, restituiamo il riferimento alla struct originale
			return &m, nil
		}
	}
	return nil, fmt.Errorf("film con ID %s non trovato", req.Id)
}
