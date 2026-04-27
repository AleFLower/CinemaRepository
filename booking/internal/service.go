package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/nats-io/nats.go"
	pb "cinema-reservation/common/proto/pb"
)

// BookingEvent definisce il messaggio che invieremo via NATS
type BookingEvent struct {
	BookingID string `json:"booking_id"`
	UserID    string `json:"user_id"`
	MovieID   string `json:"movie_id"`
	SeatID    int32  `json:"seat_id"`
}

type RoomState struct {
	Mu    sync.Mutex
	Seats map[int32]bool
}

type BookingService struct {
	pb.UnimplementedBookingServiceServer
	projections map[string]*RoomState
	mu          sync.RWMutex
	nc          *nats.Conn
}

func NewBookingService(nc *nats.Conn) *BookingService {
	return &BookingService{
		projections: make(map[string]*RoomState),
		nc:          nc,
	}
}

func (s *BookingService) getOrCreateRoom(movieID string) *RoomState {
	s.mu.RLock()
	room, exists := s.projections[movieID]
	s.mu.RUnlock()

	if exists {
		return room
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if room, exists := s.projections[movieID]; exists {
		return room
	}

	newSeats := make(map[int32]bool)
	for i := int32(1); i <= 50; i++ {
		newSeats[i] = false
	}

	newRoom := &RoomState{Seats: newSeats}
	s.projections[movieID] = newRoom
	return newRoom
}

func (s *BookingService) GetSeats(ctx context.Context, req *pb.SeatsRequest) (*pb.SeatsResponse, error) {
	room := s.getOrCreateRoom(req.MovieId)

	room.Mu.Lock()
	defer room.Mu.Unlock()

	seatsCopy := make(map[int32]bool)
	for k, v := range room.Seats {
		seatsCopy[k] = v
	}

	return &pb.SeatsResponse{
		MovieId: req.MovieId,
		Seats:   seatsCopy,
	}, nil
}

func (s *BookingService) ReserveSeat(ctx context.Context, req *pb.ReserveRequest) (*pb.ReserveResponse, error) {
	room := s.getOrCreateRoom(req.MovieId)

	strategy, err := SeatStrategyFactory(req.Strategy)
	if err != nil {
		return &pb.ReserveResponse{
			Success: false,
			Message: fmt.Sprintf("Errore strategia: %v", err),
		}, nil
	}

	room.Mu.Lock()
	defer room.Mu.Unlock()

	// VALIDAZIONE
	if err := strategy.Validate(room, req.SeatId); err != nil {
		return &pb.ReserveResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// UPDATE
	room.Seats[req.SeatId] = true
	bookingID := fmt.Sprintf("RES-%s-%d", req.MovieId, req.SeatId)

	// EVENTO
	event := BookingEvent{
		BookingID: bookingID,
		UserID:    req.UserId,
		MovieID:   req.MovieId,
		SeatID:    req.SeatId,
	}

	eventData, err := json.Marshal(event)
	if err != nil {
		log.Printf("errore serializzazione evento: %v", err)
	}

	js, err := s.nc.JetStream()
	if err == nil {
		_, err = js.Publish("bookings.created", eventData)
		if err != nil {
			log.Printf("errore invio evento: %v", err)
		}
	}

	return &pb.ReserveResponse{
		Success:   true,
		Message:   "Prenotazione effettuata con successo!",
		BookingId: bookingID,
	}, nil
}
