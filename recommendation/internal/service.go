package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strconv"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"

	pb "cinema-reservation/common/proto/pb"
)

// ==============================
// EVENT
// ==============================

type BookingEvent struct {
	BookingID    string `json:"booking_id"`
	UserID       string `json:"user_id"`
	ProjectionID string `json:"projection_id"`
	SeatID       int32  `json:"seat_id"`
}

// ==============================
// SERVICE
// ==============================

type RecommendationService struct {
	pb.UnimplementedRecommendationServiceServer
	js      nats.JetStreamContext
	redis   *redis.Client
	catalog pb.CatalogServiceClient
}

// ==============================
// CONSTRUCTOR
// ==============================

func NewRecommendationService(
	nc *nats.Conn,
	rdb *redis.Client,
	catalog pb.CatalogServiceClient,
) *RecommendationService {

	js, err := nc.JetStream()
	if err != nil {
		log.Fatalf("JetStream error: %v", err)
	}

	s := &RecommendationService{
		js:      js,
		redis:   rdb,
		catalog: catalog,
	}

	go s.consumeEvents()
	return s
}

// ==============================
// CONSUMER
// ==============================

func (s *RecommendationService) consumeEvents() {
	for {
		err := s.subscribe()
		if err != nil {
			log.Printf("[RECOMMENDATION] error: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}
		break
	}
}

func (s *RecommendationService) subscribe() error {

	_, err := s.js.QueueSubscribe(
		"bookings.event.*",
		"recommendation_group",
		func(m *nats.Msg) {

			var e BookingEvent
			if err := json.Unmarshal(m.Data, &e); err != nil {
				m.Ack()
				return
			}

			s.processEvent(e)
			m.Ack()
		},
		nats.Durable("recommendation-worker"),
		nats.ManualAck(),
		nats.DeliverAll(),
	)

	return err
}

// ==============================
// EVENT PROCESSING (SMART)
// ==============================

func (s *RecommendationService) processEvent(e BookingEvent) {

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	proj, err := s.catalog.GetProjection(ctx, &pb.ProjectionRequest{
		Id: e.ProjectionID,
	})
	if err != nil {
		log.Printf("[REC] projection error: %v", err)
		return
	}

	movie, err := s.catalog.GetMovie(ctx, &pb.MovieRequest{
		Id: proj.MovieId,
	})
	if err != nil {
		log.Printf("[REC] movie error: %v", err)
		return
	}

	key := fmt.Sprintf("user:%s:scores", e.UserID)

	pipe := s.redis.Pipeline()

	// category boost
	pipe.HIncrBy(ctx, key, "cat:"+movie.Category, 1)

	// movie preference boost (higher weight)
	pipe.HIncrBy(ctx, key, "movie:"+movie.Id, 2)

	// time slot learning (soft signal)
	pipe.HIncrBy(ctx, key, "time:"+proj.ShowTime, 1)

	_, err = pipe.Exec(ctx)
	if err != nil {
		log.Printf("[REC] redis error: %v", err)
	}
}

// ==============================
// GET RECOMMENDATIONS (RANKING)
// ==============================

type rankedMovie struct {
	Movie *pb.Movie
	Score float64
}

func (s *RecommendationService) GetRecommendations(
	ctx context.Context,
	req *pb.RecommendationRequest,
) (*pb.RecommendationResponse, error) {

	key := fmt.Sprintf("user:%s:scores", req.UserId)

	prefs, err := s.redis.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	moviesResp, err := s.catalog.GetMovies(ctx, &pb.Empty{})
	if err != nil {
		return nil, err
	}

	var ranked []rankedMovie

	for _, m := range moviesResp.Movies {

		var score float64

		for k, v := range prefs {
			val, _ := strconv.ParseFloat(v, 64)

			// category match
			if k == "cat:"+m.Category {
				score += val * 1.0
			}

			// direct movie preference boost
			if k == "movie:"+m.Id {
				score += val * 3.0
			}
		}

		ranked = append(ranked, rankedMovie{
			Movie: m,
			Score: score,
		})
	}

	sort.Slice(ranked, func(i, j int) bool {
		return ranked[i].Score > ranked[j].Score
	})

	res := make([]*pb.Movie, len(ranked))
	for i := range ranked {
		res[i] = ranked[i].Movie
	}

	return &pb.RecommendationResponse{
		Movies: res,
	}, nil
}
