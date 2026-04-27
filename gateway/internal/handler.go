package internal

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sony/gobreaker"
	pb "cinema-reservation/common/proto/pb"
)

type GatewayHandler struct {
	catalogClient pb.CatalogServiceClient
	bookingClients []pb.BookingServiceClient // Slice per lo sharding
	cb            *gobreaker.CircuitBreaker
}

func NewGatewayHandler(catClient pb.CatalogServiceClient, bookClients []pb.BookingServiceClient) *GatewayHandler {
	// Configurazione Circuit Breaker
	settings := gobreaker.Settings{
		Name:        "BookingService",
		MaxRequests: 3,
		Interval:    5 * time.Second,
		Timeout:     10 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures > 2
		},
	}

	return &GatewayHandler{
		catalogClient:  catClient,
		bookingClients: bookClients,
		cb:             gobreaker.NewCircuitBreaker(settings),
	}
}

// GetMovies: Aggrega i dati dal Catalog Service
func (h *GatewayHandler) GetMovies(c *gin.Context) {
	resp, err := h.catalogClient.GetMovies(context.Background(), &pb.Empty{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ReserveSeat: Implementa Circuit Breaker e Sharding
func (h *GatewayHandler) ReserveSeat(c *gin.Context) {
	var req struct {
		MovieID  string `json:"movie_id" binding:"required"`
		SeatID   int32  `json:"seat_id"`
		UserID   string `json:"user_id" binding:"required"`
		Strategy string `json:"strategy"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 1. LOGICA DI SHARDING (Semplice Modulo sull'ultimo carattere dell'ID)
	// In Docker Compose abbiamo 2 repliche: booking-service:50051 e 50052
	shardIndex := int(req.MovieID[len(req.MovieID)-1]) % len(h.bookingClients)
	targetClient := h.bookingClients[shardIndex]

	// 2. CIRCUIT BREAKER
	result, err := h.cb.Execute(func() (interface{}, error) {
		return targetClient.ReserveSeat(context.Background(), &pb.ReserveRequest{
			MovieId:  req.MovieID,
			SeatId:   req.SeatID,
			UserId:   req.UserID,
			Strategy: req.Strategy,
		})
	})

	if err != nil {
		if err == gobreaker.ErrOpenState {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Il servizio prenotazioni è temporaneamente fuori servizio (CB Open)"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
