package internal

import (
	"encoding/json"
	"log"

	"github.com/nats-io/nats.go"
)

type BookingEvent struct {
	BookingID string `json:"booking_id"`
	UserID    string `json:"user_id"`
	ProjectionID   string `json:"projection_id"`
	SeatID    int32  `json:"seat_id"`
}

type NotificationService struct {
	nc *nats.Conn
}

func NewNotificationService(nc *nats.Conn) *NotificationService {
	return &NotificationService{nc: nc}
}

func (s *NotificationService) Start() {
	js, err := s.nc.JetStream()
	if err != nil {
		log.Fatalf("JetStream error: %v", err)
	}

	// Queue subscription (load-balanced if you scale notification service)
	// Listens to booking event subject: bookings.event.*
	_, err = js.QueueSubscribe("bookings.event.*", "notification_group", func(m *nats.Msg) {
		var event BookingEvent
		if err := json.Unmarshal(m.Data, &event); err != nil {
		       log.Printf("Error: %s", err)
			m.Ack() // Ignore malformed messages
			return
		}

		s.sendEmail(event)
		errAck := m.Ack()
    if errAck != nil {
        log.Printf("[ACK ERROR] %v", errAck)
    }
	},
		nats.Durable("notification-worker"),
		nats.ManualAck(),
		nats.DeliverAll(),
	)

	if err != nil {
		log.Fatalf("Subscription error: %v", err)
	}

	log.Println("[NOTIFICATION] Ready, listening for new events...")
}

func (s *NotificationService) sendEmail(e BookingEvent) {
	log.Printf("[EMAIL SENT] Confirmation for %s: Seat %d successfully booked!", e.UserID, e.SeatID)
}
