package internal

import (
	"encoding/json"
	"log"

	"github.com/nats-io/nats.go"
)

type BookingEvent struct {
	BookingID string `json:"booking_id"`
	UserID    string `json:"user_id"`
	MovieID   string `json:"movie_id"`
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

	// Sottoscrizione Queue (bilanciata se scali il servizio notifiche)
	// Ascolta lo stesso subject del booking: bookings.event.*
	_, err = js.QueueSubscribe("bookings.event.*", "notification_group", func(m *nats.Msg) {
		var event BookingEvent
		if err := json.Unmarshal(m.Data, &event); err != nil {
			m.Ack() // Ignora messaggi malformati
			return
		}

		s.sendEmail(event)
		m.Ack() // Conferma ricezione
	}, 
	nats.Durable("notification-worker"), 
	nats.ManualAck(),
	nats.DeliverAll(),
	)

	if err != nil {
		log.Fatalf("Sub error: %v", err)
	}

	log.Println("[NOTIFICATION] Pronto, in ascolto di nuovi eventi...")
}

func (s *NotificationService) sendEmail(e BookingEvent) {
	log.Printf("📧 [EMAIL SENT] Conferma per %s: Posto %d prenotato con successo!", e.UserID, e.SeatID)
}
