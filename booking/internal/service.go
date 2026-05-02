package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"
	"os"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
	

	pb "cinema-reservation/common/proto/pb"
	"cinema-reservation/common/utils"
)

type BookingEvent struct {
	BookingID string `json:"booking_id"`
	UserID    string `json:"user_id"`
	ProjectionID   string `json:"projection_id"`
	SeatID    int32  `json:"seat_id"`
}

type RoomState struct {
	Mu    sync.Mutex
	Seats map[int32]bool
	Strategy string
}

type BookingService struct {
	pb.UnimplementedBookingServiceServer
	projections map[string]*RoomState
	nc          *nats.Conn
	js          nats.JetStreamContext
	redis       *redis.Client
}

var unlockScript = redis.NewScript(`
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("DEL", KEYS[1])
else
	return 0
end
`)

func NewBookingService(nc *nats.Conn, rdb *redis.Client, configPath string) *BookingService {
	js, err := nc.JetStream()
	if err != nil {
		log.Fatalf("JetStream error: %v", err)
	}

	streamName := utils.GetEnv("NATS_STREAM", "BOOKINGS_STREAM")
	subjectAll := utils.GetEnv("NATS_SUBJECT_ALL", "bookings.>")
	subjectEvents := utils.GetEnv("NATS_SUBJECT_EVENTS", "bookings.event.*")

	_, err = js.AddStream(&nats.StreamConfig{
		Name:     streamName,
		Subjects: []string{subjectAll},
		Storage:  nats.FileStorage,
	})
	if err != nil {
		log.Printf("Stream già esistente: %v", err)
	}

	file, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalf("Errore lettura config: %v", err)
	}

	var config []struct {
		ID         string `json:"id"`
		TotalSeats int    `json:"total_seats"`
		Strategy   string `json:"strategy"`
	}

	if err := json.Unmarshal(file, &config); err != nil {
		log.Fatalf("Errore parsing config: %v", err)
	}

	projections := make(map[string]*RoomState)

	for _, p := range config {
	seats := make(map[int32]bool)

	for i := int32(1); i <= int32(p.TotalSeats); i++ {
		seats[i] = false
	}

	strategy := p.Strategy
	if strategy == "" {
		strategy = utils.GetEnv("DEFAULT_STRATEGY", "default")
	}

	projections[p.ID] = &RoomState{
		Seats:    seats,
		Strategy: strategy,
	}
}

	s := &BookingService{
		projections: projections,
		nc:          nc,
		js:          js,
		redis:       rdb,
	}

	s.restoreState(subjectEvents)
	go s.syncLiveEvents(subjectEvents)

	return s
}

func (s *BookingService) restoreState(subject string) {
	timeout := utils.GetEnvDuration("RESTORE_TIMEOUT", 500*time.Millisecond)

	sub, err := s.js.SubscribeSync(subject, nats.DeliverAll())
	if err != nil {
		return
	}

	log.Println("[EVENT SOURCING] restore state...")

	for {
		msg, err := sub.NextMsg(timeout)
		if err != nil {
			break
		}

		var event BookingEvent
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			continue
		}

		if room, ok := s.projections[event.ProjectionID]; ok {
			room.Seats[event.SeatID] = true
		}
	}
}

func (s *BookingService) syncLiveEvents(subject string) {
	s.js.Subscribe(subject, func(m *nats.Msg) {
		var event BookingEvent
		if err := json.Unmarshal(m.Data, &event); err != nil {
			return
		}

		if room, ok := s.projections[event.ProjectionID]; ok {
			room.Mu.Lock()
			room.Seats[event.SeatID] = true
			room.Mu.Unlock()
		}
	}, nats.DeliverNew())
}

func (s *BookingService) acquireLock(ctx context.Context, key string) (string, bool, error) {
	token := uuid.NewString()

	ttl := utils.GetEnvDuration("LOCK_TTL", 5*time.Second)

	ok, err := s.redis.SetNX(ctx, key, token, ttl).Result()
	if err != nil {
		return "", false, err
	}

	return token, ok, nil
}

func (s *BookingService) releaseLock(ctx context.Context, key, token string) {
	unlockScript.Run(ctx, s.redis, []string{key}, token)
}

func (s *BookingService) ReserveSeat(ctx context.Context, req *pb.ReserveRequest) (*pb.ReserveResponse, error) {

	log.Printf("Prova")

	room, err := s.getRoom(req.ProjectionId)
	if err != nil {
		return nil, err
	}
	
	if _, exists := room.Seats[req.SeatId]; !exists {
          return &pb.ReserveResponse{
          Success: false,
          Message: "posto non esistente",
            }, nil
        }

	lockKey := fmt.Sprintf("lock:seat:%s:%d", req.ProjectionId, req.SeatId)

	token, ok, err := s.acquireLock(ctx, lockKey)
	if err != nil {
		return nil, fmt.Errorf("redis error: %v", err)
	}

	if !ok {
		return &pb.ReserveResponse{
			Success: false,
			Message: "posto in fase di prenotazione, riprova",
		}, nil
	}

	defer s.releaseLock(ctx, lockKey, token)

	strategy, _ := SeatStrategyFactory(room.Strategy)

	room.Mu.Lock()
	defer room.Mu.Unlock()

	if err := strategy.Validate(room, req.SeatId); err != nil {
		return &pb.ReserveResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	if room.Seats[req.SeatId] {
		return &pb.ReserveResponse{
			Success: false,
			Message: "posto già occupato",
		}, nil
	}

	event := BookingEvent{
		BookingID: fmt.Sprintf("RES-%s-%d", req.ProjectionId, req.SeatId),
		UserID:    req.UserId,
		ProjectionID:   req.ProjectionId,
		SeatID:    req.SeatId,
	}

	data, _ := json.Marshal(event)

	subjectPrefix := utils.GetEnv("NATS_SUBJECT_PREFIX", "bookings.event")
	subject := fmt.Sprintf("%s.%s", subjectPrefix, req.ProjectionId)

	_, err = s.js.Publish(subject, data)
	if err != nil {
		return nil, fmt.Errorf("errore publish event: %v", err)
	}

	room.Seats[req.SeatId] = true

	return &pb.ReserveResponse{
		Success:   true,
		Message:   "Prenotazione confermata",
		BookingId: event.BookingID,
	}, nil
}

func (s *BookingService) GetSeats(ctx context.Context, req *pb.SeatsRequest) (*pb.SeatsResponse, error) {
	room, err := s.getRoom(req.ProjectionId)
	if err != nil {
		return nil, err
	}

	room.Mu.Lock()
	defer room.Mu.Unlock()

	copySeats := make(map[int32]bool)
	for k, v := range room.Seats {
		copySeats[k] = v
	}

	return &pb.SeatsResponse{
		ProjectionId: req.ProjectionId,
		Seats:   copySeats,
	}, nil
}

func (s *BookingService) getRoom(id string) (*RoomState, error) {
	if r, ok := s.projections[id]; ok {
		return r, nil
	}
	return nil, fmt.Errorf("proiezione non trovata")
}
