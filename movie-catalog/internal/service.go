package internal

import (
	"context"
	"fmt"

	pb "cinema-reservation/common/proto/pb"
)

type CatalogService struct {
	pb.UnimplementedCatalogServiceServer
	movies []pb.Movie
}

func NewCatalogService() *CatalogService {
	// Dati di esempio (Stateless: non cambiano durante l'esecuzione)
	return &CatalogService{
		movies: []pb.Movie{
			{Id: "101", Title: "Dune: Part Two", ShowTime: "20:30", TotalSeats: 50},
			{Id: "102", Title: "Oppenheimer", ShowTime: "21:00", TotalSeats: 50},
			{Id: "103", Title: "Interstellar", ShowTime: "18:00", TotalSeats: 50},
		},
	}
}

func (s *CatalogService) GetMovies(ctx context.Context, req *pb.Empty) (*pb.MovieList, error) {
	fmt.Println("Richiesta lista film ricevuta")
	// Trasformiamo lo slice di struct in un puntatore a MovieList
	moviePtrs := make([]*pb.Movie, len(s.movies))
	for i := range s.movies {
		moviePtrs[i] = &s.movies[i]
	}
	return &pb.MovieList{Movies: moviePtrs}, nil
}

func (s *CatalogService) GetMovie(ctx context.Context, req *pb.MovieRequest) (*pb.Movie, error) {
	for _, m := range s.movies {
		if m.Id == req.Id {
			return &m, nil
		}
	}
	return nil, fmt.Errorf("film con ID %s non trovato", req.Id)
}
