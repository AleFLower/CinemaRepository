package internal

import (
	"context"
	"net/http"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sony/gobreaker"

	pb "cinema-reservation/common/proto/pb"
)

type GatewayHandler struct {
	catalogClient pb.CatalogServiceClient
	bookingClient pb.BookingServiceClient
	cb            *gobreaker.CircuitBreaker
}

func NewGatewayHandler(
	catClient pb.CatalogServiceClient,
	bookClient pb.BookingServiceClient,
) *GatewayHandler {

	settings := gobreaker.Settings{
		Name:        "BookingService",
		MaxRequests: 5,
		Interval:    10 * time.Second,
		Timeout:     15 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 3
		},
	}

	return &GatewayHandler{
		catalogClient: catClient,
		bookingClient: bookClient,
		cb:            gobreaker.NewCircuitBreaker(settings),
	}
}

//
// GET /movies
//
func (h *GatewayHandler) GetMovies(c *gin.Context) {
	resp, err := h.catalogClient.GetMovies(context.Background(), &pb.Empty{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Catalog service non disponibile",
			"details": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, resp)
}

//
// GET /seats/:id
//
func (h *GatewayHandler) GetSeats(c *gin.Context) {
	projectionID := c.Param("id")

	resp, err := h.bookingClient.GetSeats(
		context.Background(),
		&pb.SeatsRequest{ProjectionId: projectionID},
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Booking service non disponibile",
			"details": err.Error(),
		})
		return
	}

	//  Ordiniamo i posti
	keys := make([]int, 0, len(resp.Seats))
	for k := range resp.Seats {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)

	seats := make([]gin.H, 0, len(keys))
	for _, k := range keys {
		seats = append(seats, gin.H{
			"seat_id":  k,
			"occupied": resp.Seats[int32(k)],
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"projection_id": resp.ProjectionId,
		"seats":    seats,
	})
}

//
// POST /book
//
func (h *GatewayHandler) ReserveSeat(c *gin.Context) {
	var req struct {
		ProjectionID string `json:"projection_id" binding:"required"`
		SeatID       int32  `json:"seat_id" binding:"required"`
		UserID       string `json:"user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Dati non validi",
			"details": err.Error(),
		})
		return
	}

	result, err := h.cb.Execute(func() (interface{}, error) {
		return h.bookingClient.ReserveSeat(
			context.Background(),
			&pb.ReserveRequest{
				ProjectionId:  req.ProjectionID,
				SeatId:   req.SeatID,
				UserId:   req.UserID,
			},
		)
	})

	if err != nil {
		if err == gobreaker.ErrOpenState {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "Servizio booking temporaneamente non disponibile (circuit breaker aperto)",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

