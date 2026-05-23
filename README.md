Ecco una versione **pulita, professionale e pronta per GitHub/portfolio**, senza eccesso di emoji e con struttura più “real-world”.

---

```markdown
# Cinema Reservation Platform

A distributed microservices system for a cinema booking platform built with Go, gRPC, Gin, Redis, NATS, and Docker Compose.

The project simulates a real-world event-driven architecture with booking, catalog, recommendations, and a CLI client.

---

## Features

- Movie catalog service (gRPC)
- Seat reservation system with concurrency handling
- Redis for caching, locking, and performance optimization
- NATS JetStream for event-driven communication
- Notification service (event consumer)
- Recommendation service (user-based suggestions)
- API Gateway (REST + Gin + gRPC clients)
- Interactive CLI client (terminal-based UI)

---

## Architecture

```

CLI → API Gateway → gRPC Services
│
┌──────────────┼────────────────┐
│              │                │
Catalog     Booking      Recommendation
Service      Service          Service
│              │                │
└────── Redis + NATS Event Bus ─┘

````

---

## Requirements

- Docker
- Docker Compose
- (Optional) Go 1.21+ for running CLI locally

---

## Setup & Run (Local)

### Clone the repository

```bash
git clone https://github.com/<your-username>/<repo-name>.git
cd <repo-name>
````

### Start all services

```bash
docker compose up --build
```

Or in detached mode:

```bash
docker compose up -d --build
```

### Stop services

```bash
docker compose down
```

To remove volumes:

```bash
docker compose down -v
```

---

## Live Deployment

The system is deployed and accessible via:

API Gateway:
[http://3.232.139.1:8080](http://3.232.139.1:8080)

---

## Environment Configuration

The API Gateway URL can be configured using an environment variable:

```bash
API_BASE_URL=http://3.232.139.1:8080
```

For local development:

```bash
API_BASE_URL=http://localhost:8080
```

---

## API Endpoints

### Get Movies

```
GET /movies
```

### Get Projections

```
GET /projections
GET /projections/:movieId
```

### Get Seats

```
GET /seats/:projectionId
```

### Book a Seat

```
POST /book
```

Body:

```json
{
  "projection_id": "1",
  "seat_id": 10,
  "user_id": "user1"
}
```

### Get Recommendations

```
GET /recommendations?user_id=user1
```

---

## CLI Client

The project includes a CLI client with a terminal-based UI.

### Run locally

```bash
cd cli
go run main.go
```

Or build:

```bash
go build -o cineflix-cli
./cineflix-cli
```

---

## Booking Flow

1. User selects a movie
2. Views available projections
3. Checks seat availability
4. Sends booking request to API Gateway
5. Booking service:

   * Validates request
   * Uses Redis for distributed locking
   * Publishes event to NATS
6. Notification and Recommendation services react asynchronously

---

## Event-Driven Architecture

The system uses NATS JetStream for asynchronous communication.

### Subjects

* `bookings.event.*`
* `bookings.>`

### Used for

* Seat reservation events
* Notifications
* Recommendation updates
* Service integration

---

## Technologies Used

* Go (Golang)
* gRPC
* Gin Web Framework
* Redis
* NATS JetStream
* Docker & Docker Compose
* Circuit Breaker (Sony gobreaker)

---

## Debug & Logs

### View logs

```bash
docker compose logs -f
```

### Full rebuild

```bash
docker compose down -v
docker compose up --build
```

---

## Notes

* API Gateway runs on port 8080
* Internal communication uses gRPC over Docker network
* All services are containerized
* Designed for distributed systems learning

---

## Project Goals

This project demonstrates:

* Microservices architecture
* Event-driven design
* gRPC communication patterns
* Fault tolerance and resilience (circuit breaker)
* Containerized deployment with Docker
* Real-world backend system design

---

## License

This project is for educational purposes only.

```

---

Se vuoi, nel prossimo step posso anche:
- :contentReference[oaicite:0]{index=0}
- :contentReference[oaicite:1]{index=1}
- oppure :contentReference[oaicite:2]{index=2}
```
