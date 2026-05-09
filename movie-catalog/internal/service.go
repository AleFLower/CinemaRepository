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
	movies      []pb.Movie
	projections []pb.Projection
}

func NewCatalogService(configPath string) *CatalogService {
	file, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalf("Error loading file: %v", err)
	}

	// Local structs matching JSON exactly
	type jsonMovie struct {
		Id       string `json:"id"`
		Title    string `json:"title"`
		Category string `json:"category"`
	}

	type jsonProjection struct {
		Id         string `json:"id"`
		MovieId    string `json:"movie_id"`
		ShowTime   string `json:"show_time"`
		RoomId     string `json:"room_id"`
		TotalSeats int32  `json:"total_seats"`
	}

	var data struct {
		Movies      []jsonMovie      `json:"movies"`
		Projections []jsonProjection `json:"projections"`
	}

	if err := json.Unmarshal(file, &data); err != nil {
		log.Fatalf("Parsing error: %v", err)
	}

	svc := &CatalogService{}

	for _, m := range data.Movies {
		svc.movies = append(svc.movies, pb.Movie{
			Id:       m.Id,
			Title:    m.Title,
			Category: m.Category,
		})
	}

	for _, p := range data.Projections {
		svc.projections = append(svc.projections, pb.Projection{
			Id:         p.Id,
			MovieId:    p.MovieId,
			ShowTime:   p.ShowTime,
			RoomId:     p.RoomId,
			TotalSeats: p.TotalSeats,
		})
	}

	log.Printf("Catalog loaded: %d movies, %d projections", len(svc.movies), len(svc.projections))
	return svc
}

//
// 🎬 MOVIES
//
func (s *CatalogService) GetMovies(ctx context.Context, req *pb.Empty) (*pb.MovieList, error) {

	log.Println("Catalog: GetMovies request")

	moviePtrs := make([]*pb.Movie, len(s.movies))
	for i := range s.movies {
		moviePtrs[i] = &s.movies[i]
	}

	return &pb.MovieList{
		Movies: moviePtrs,
	}, nil
}

//
// 🎬 SINGLE MOVIE
//
func (s *CatalogService) GetMovie(ctx context.Context, req *pb.MovieRequest) (*pb.Movie, error) {

	for i := range s.movies {
		if s.movies[i].Id == req.Id {
			return &s.movies[i], nil
		}
	}

	return nil, fmt.Errorf("movie with ID %s not found", req.Id)
}

//
// 🎟️ ALL PROJECTIONS
//
func (s *CatalogService) GetProjections(ctx context.Context, req *pb.Empty) (*pb.ProjectionList, error) {

	res := make([]*pb.Projection, len(s.projections))

	for i := range s.projections {
		res[i] = &s.projections[i]
	}

	return &pb.ProjectionList{
		Projections: res,
	}, nil
}

//
// 🎟️ PROJECTIONS BY MOVIE
//
func (s *CatalogService) GetProjectionsByMovie(ctx context.Context, req *pb.MovieRequest) (*pb.ProjectionList, error) {

	var result []*pb.Projection

	for i := range s.projections {
		if s.projections[i].MovieId == req.Id {
			result = append(result, &s.projections[i])
		}
	}

	return &pb.ProjectionList{
		Projections: result,
	}, nil
}

func (s *CatalogService) GetProjection(ctx context.Context, req *pb.ProjectionRequest) (*pb.Projection, error) {

	for i := range s.projections {
		if s.projections[i].Id == req.Id {
			return &s.projections[i], nil
		}
	}

	return nil, fmt.Errorf("projection %s not found", req.Id)
}
