package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nats-io/nats.go"
	"cinema-reservation/notification/internal"
)

func main() {
	// 1. Connessione a NATS
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = nats.DefaultURL
	}

	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatalf("Impossibile connettersi a NATS: %v", err)
	}
	defer nc.Close()

	// 2. Avvio del servizio
	service := internal.NewNotificationService(nc)
	service.Start()

	// 3. Keep-alive: Aspetta un segnale di interruzione (Ctrl+C)
	log.Println("Notification Service avviato correttamente.")
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
}
