package internal

import (
	"context"
	"net/http"
	"sort"
	"time"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/sony/gobreaker"
	"cinema-reservation/common/utils"

	pb "cinema-reservation/common/proto/pb"
)

type GatewayHandler struct {
	catalogClient        pb.CatalogServiceClient
	bookingClient        pb.BookingServiceClient
	cb                   *gobreaker.CircuitBreaker
	recommendationClient pb.RecommendationServiceClient
}

func NewGatewayHandler(
	catClient pb.CatalogServiceClient,
	bookClient pb.BookingServiceClient,
	recClient pb.RecommendationServiceClient,
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
		catalogClient:        catClient,
		bookingClient:        bookClient,
		recommendationClient: recClient,
		cb:                   gobreaker.NewCircuitBreaker(settings),
	}
}

//
// 🎬 GET /movies
//
func (h *GatewayHandler) GetMovies(c *gin.Context) {
	resp, err := h.catalogClient.GetMovies(context.Background(), &pb.Empty{})
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": utils.MapGRPCErrorToUser(err),
		})
		return
	}
	c.JSON(http.StatusOK, resp)
}

//
// 💺 GET /seats/:id
//
func (h *GatewayHandler) GetSeats(c *gin.Context) {
	projectionID := c.Param("id")

	resp, err := h.bookingClient.GetSeats(
		context.Background(),
		&pb.SeatsRequest{ProjectionId: projectionID},
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": utils.MapGRPCErrorToUser(err),
		})
		return
	}

	// Sort seats
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
		"seats":         seats,
	})
}

//
// 🎟️ POST /book
//
func (h *GatewayHandler) ReserveSeat(c *gin.Context) {
	var req struct {
		ProjectionID string `json:"projection_id" binding:"required"`
		SeatID       int32  `json:"seat_id" binding:"required"`
		UserID       string `json:"user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "The provided data is not valid.",
		})
		return
	}

	result, err := h.cb.Execute(func() (interface{}, error) {

		return h.bookingClient.ReserveSeat(
			context.Background(),
			&pb.ReserveRequest{
				ProjectionId: req.ProjectionID,
				SeatId:       req.SeatID,
				UserId:       req.UserID,
			},
		)
	})

	if err != nil {
		// Circuit Breaker handling
		if err == gobreaker.ErrOpenState {
			log.Println("retry")
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "The booking system is temporarily overloaded. Please try again in a few seconds.",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": utils.MapGRPCErrorToUser(err),
		})
		return
	}

	res := result.(*pb.ReserveResponse)
	if !res.Success {
		c.JSON(http.StatusConflict, gin.H{
			"error": res.Message,
		})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *GatewayHandler) GetRecommendations(c *gin.Context) {
	userID := c.Query("user_id")

	resp, err := h.recommendationClient.GetRecommendations(
		context.Background(),
		&pb.RecommendationRequest{UserId: userID},
	)

	if err != nil {
		c.JSON(500, gin.H{"error": "recommendation error"})
		return
	}

	c.JSON(200, resp)
}

func (h *GatewayHandler) GetProjections(c *gin.Context) {
	resp, err := h.catalogClient.GetProjections(context.Background(), &pb.Empty{})
	if err != nil {
		c.JSON(500, gin.H{"error": "catalog error"})
		return
	}
	c.JSON(200, resp)
}

func (h *GatewayHandler) GetProjectionsByMovie(c *gin.Context) {
	movieID := c.Param("id")

	resp, err := h.catalogClient.GetProjectionsByMovie(
		context.Background(),
		&pb.MovieRequest{Id: movieID},
	)

	if err != nil {
		c.JSON(500, gin.H{"error": "catalog error"})
		return
	}

	c.JSON(200, resp)
}
