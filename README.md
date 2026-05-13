
# 🎬 Cinema Reservation Platform

A distributed microservices system for a cinema booking platform built with **Go, gRPC, Gin, Redis, NATS, and Docker Compose**.

The project simulates a real-world event-driven architecture with booking, catalog, recommendations, and a CLI client.

---

## 🚀 Features

- 🎬 Movie catalog service (gRPC)
- 🎟️ Seat reservation system with concurrency handling
- 🧠 Redis for caching, locking, and performance
- 📡 NATS for event-driven communication
- 🔔 Notification service (event consumer)
- 🎯 Recommendation service (user-based suggestions)
- 🌐 API Gateway (REST + Gin + gRPC clients)
- 🖥️ Interactive CLI client (Netflix-style terminal UI)

---

## 🧱 Architecture

```

CLI → API Gateway → gRPC Services
│
┌───────────────┼────────────────┐
│               │                │
Catalog Service   Booking Service   Recommendation Service
│               │                │
└─────── Redis + NATS Event Bus ─┘

````

---

## 📦 Requirements

Make sure you have installed:

- Docker
- Docker Compose
- (Optional) Go 1.21+ (for running CLI locally)

---

## ⚙️ Setup & Run

### 1. Clone the repository

```bash
git clone https://github.com/<your-username>/<repo-name>.git
cd <repo-name>
````

---

### 2. Start all services

```bash
docker compose up --build
```

Or in detached mode:

```bash
docker compose up -d --build
```

---

### 3. Stop everything

```bash
docker compose down
```

To also remove volumes:

```bash
docker compose down -v
```

---

## 🌐 Services

After startup, services are available at:

| Service            | URL                                            |
| ------------------ | ---------------------------------------------- |
| 🌐 API Gateway     | [http://localhost:8080](http://localhost:8080) |
| 📡 NATS Monitoring | [http://localhost:8222](http://localhost:8222) |
| 🧠 Redis           | localhost:6379                                 |

---

## 🎬 API Endpoints

### 📌 Get Movies

```http
GET /movies
```

---

### 📌 Get Projections

```http
GET /projections
GET /projections/:movieId
```

---

### 📌 Get Seats

```http
GET /seats/:projectionId
```

---

### 📌 Book a Seat

```http
POST /book
Content-Type: application/json

{
  "projection_id": "1",
  "seat_id": 10,
  "user_id": "user1"
}
```

---

### 📌 Get Recommendations

```http
GET /recommendations?user_id=user1
```

---

## 🖥️ CLI Client (CineFlix Terminal UI)

The project includes an interactive CLI client inspired by Netflix UI.

### Run CLI locally

```bash
cd cli
go run main.go
```

Or build it:

```bash
go build -o cineflix-cli
./cineflix-cli
```

---

## 🎟️ Booking Flow

1. User selects a movie
2. Views available projections
3. Checks seat availability
4. Books a seat via API Gateway
5. Booking service:

   * Validates request
   * Uses Redis for locking
   * Publishes event to NATS
6. Notification & recommendation services react to events

---

## 📡 Event-Driven Architecture

The system uses **NATS JetStream** for messaging.

### Subjects:

* `bookings.event.*`
* `bookings.>`

### Used for:

* Seat reservation events
* Notifications
* Recommendation updates
* System integration

---

## 🧠 Technologies Used

* Go (Golang)
* gRPC
* Gin Web Framework
* Redis
* NATS JetStream
* Docker & Docker Compose
* Circuit Breaker (Sony gobreaker)

---

## 🧪 Debug & Logs

### View logs

```bash
docker compose logs -f
```

### Rebuild everything cleanly

```bash
docker compose down -v
docker compose up --build
```

---

## ⚠️ Notes

* API Gateway runs on port **8080**
* Internal communication uses **gRPC over Docker network**
* Services are fully containerized
* Designed for learning distributed systems & microservices

---

## 📌 Project Goal

This project is built to demonstrate:

* Microservices design
* Event-driven architecture
* gRPC communication
* Fault tolerance (circuit breaker)
* Containerized deployment
* Real-world backend system design

---

## 📜 License

This project is for educational purposes only.
