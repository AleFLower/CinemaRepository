package internal

import (
	"encoding/json"
	"log"

	"github.com/nats-io/nats.go"
)

// BookingEvent deve corrispondere alla struttura definita nel Booking Service
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

// Start ascolta i messaggi e definisce la logica di reazione
func (s *NotificationService) Start() {
    js, err := s.nc.JetStream()
    if err != nil {
        log.Fatalf("Errore JetStream: %v", err)
    }

    // Assicuriamoci che lo stream esista
    _, err = js.AddStream(&nats.StreamConfig{
        Name:     "BOOKINGS_STREAM",
        Subjects: []string{"bookings.created"},
        Storage:  nats.FileStorage,
    })

    // Sottoscrizione DURABLE con DELIVER ALL
    _, err = js.QueueSubscribe("bookings.created", "notification_group", func(m *nats.Msg) {
        var event BookingEvent
        if err := json.Unmarshal(m.Data, &event); err != nil {
            return
		}
        s.sendEmail(event)
        m.Ack()
    }, 
    nats.Durable("worker-notification"), // Nome unico per ricordare la posizione
    nats.ManualAck(),                    // Conferma manuale per sicurezza
    nats.DeliverAll(),                   // <--- QUESTA RECUPERA I MESSAGGI PASSATI
    )

    if err != nil {
        log.Fatalf("Errore sottoscrizione: %v", err)
    }

    log.Println("Notification Service pronto e in ascolto dello storico...")
}

func (s *NotificationService) sendEmail(e BookingEvent) {
	// Qui simuli l'invio dell'email
	log.Printf(" [EMAIL SENT] Gentile utente %s, la tua prenotazione %s per il film %s (Posto: %d) è confermata!", 
		e.UserID, e.BookingID, e.MovieID, e.SeatID)
}
