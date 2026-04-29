package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"os"

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

func NewBookingService(nc *nats.Conn, configPath string) *BookingService {
    file, err := os.ReadFile(configPath)
    if err != nil {
        log.Fatalf("Errore config: %v", err)
    }

    var config []struct {
        ID         string `json:"id"`
        TotalSeats int    `json:"total_seats"`
    }
    json.Unmarshal(file, &config)

    projections := make(map[string]*RoomState)
    for _, p := range config {
        seats := make(map[int32]bool)
        for i := int32(1); i <= int32(p.TotalSeats); i++ {
            seats[i] = false
        }
        projections[p.ID] = &RoomState{Seats: seats}
        log.Printf("[INIT] Proiezione %s caricata (%d posti)", p.ID, p.TotalSeats)
    }

    return &BookingService{projections: projections, nc: nc}
}

func (s *BookingService) getRoom(movieID string) (*RoomState, error) {
    // Non serve Lock su s.projections perché la mappa non viene più modificata dopo il New
    room, exists := s.projections[movieID]
    if !exists {
        return nil, fmt.Errorf("film con ID %s non disponibile o sala inesistente", movieID)
    }
    return room, nil
}

func (s *BookingService) GetSeats(ctx context.Context, req *pb.SeatsRequest) (*pb.SeatsResponse, error) {
	room, err := s.getRoom(req.MovieId)
        if err != nil {
         return nil, err // Ritorna errore gRPC
        }

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
   log.Printf("[BOOKING SERVICE] Ricevuta richiesta per Proiezione: %s, Posto: %d", req.MovieId, req.SeatId)
    // 1. Recupero la proiezione specifica
    room, err := s.getRoom(req.MovieId)
    if err != nil {
        return &pb.ReserveResponse{Success: false, Message: err.Error()}, nil
    }

    // 2. Validazione Range Posti (dinamica sulla sala della proiezione)
    if req.SeatId < 1 || req.SeatId > int32(len(room.Seats)) {
        return &pb.ReserveResponse{
            Success: false,
            Message: fmt.Sprintf("Posto non valido per questa sala (Max: %d)", len(room.Seats)),
        }, nil
    }

    // 3. Factory Strategia
    strategy, _ := SeatStrategyFactory(req.Strategy)

    room.Mu.Lock()
    defer room.Mu.Unlock()

    // 4. Validazione Algoritmica (NoGap/Social)
    if err := strategy.Validate(room, req.SeatId); err != nil {
        return &pb.ReserveResponse{Success: false, Message: err.Error()}, nil
    }

    // 5. Occupazione e Evento
    room.Seats[req.SeatId] = true
    bookingID := fmt.Sprintf("RES-%s-%d", req.MovieId, req.SeatId)

    event := BookingEvent{
        BookingID: bookingID,
        UserID:    req.UserId,
        MovieID:   req.MovieId, // Qui rappresenta l'ID proiezione
        SeatID:    req.SeatId,
    }
    
    eventData, _ := json.Marshal(event)
    js, _ := s.nc.JetStream()
    js.Publish("bookings.created", eventData)

    return &pb.ReserveResponse{
        Success: true,
        Message: "Prenotazione confermata",
        BookingId: bookingID,
    }, nil
}
