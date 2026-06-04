
# 🎬 Cinema reservation microservices application

This is an event-driven cinema reservation platform built using Go, gRPC, and the Gin Web Framework. 

This project features an interactive CLI Client, an API Gateway with built-in resilience, and multiple distributed backend microservices communicating via gRPC and event streams.

---

## 🏗️ System Architecture

The platform is split into decoupled, specialized components:

* **CLI Client (`main.go`):** A command-line user interface wrapped with clean terminal styling for browsing listings, inspecting seating maps, booking seats, and tracking user profiles.
* **API Gateway:** Translates incoming HTTP REST requests into internal gRPC calls. It implements client-side load balancing (`round_robin`) to distribute traffic and wraps critical paths (like reservations) with a **Circuit Breaker** (`gobreaker`) to isolate microservice failures.
* **Movie Catalog Service:** Manages films and historical schedule listings.
* **Booking Service:** Manages active seat maps. Utilizes **Redis** for distributed locking/atomic operations and dispatches streams to **NATS JetStream**.
* **Recommendation & Notification Services:** Asynchronous, event-driven services listening to broker topics to handle system events and user preferences.

---

## 🚀 Getting Started

### Prerequisites
* **Go** (v1.18 or higher) installed.
* **Docker** and **Docker Compose** installed (only required for local cluster deployment).

---

### 🐳 Run the Entire System Locally

To spin up the complete isolated microservices environment on your local machine using Docker Compose:

#### 1. Update the Client Configuration

Open the CLI client file (main.go) and ensure that the baseURL points to your local machine:

```go
const baseURL = "http://localhost:8080"

```

#### 2. Start the Docker Containers

From the root directory containing your `docker-compose.yml`, run:

```bash
docker compose up --build

```


This command builds and runs the following environment:

* **nats-broker** (Ports `4222`, `8222`)
* **redis** (Port `6379`)
* **gateway** (Port `8080`)
* **catalog-service**, **booking**, **notification-service**, **recommendation-service**

Launch & Scale Infrastructure: Open a terminal in the root directory of the project. To test how the API Gateway balances traffic across multiple instances of a service using the internal round-robin gRPC balancer, you can scale the backend nodes dynamically:
```bash

   docker compose up -d --scale cinema-catalog-service=3 --scale cinema-booking-service=2
```

🧹 How to Clean and Restart Everything From Scratch
```bash
docker compose down -v
```

#### 3. Live demo: Run the CLI Client

Open a new, separate terminal tab or window and execute:

```bash
go run main.go

```

---

## 🛠️ API REST Reference

The API Gateway exposes the following REST matrix for external integrations, cURL, or Postman:

| HTTP Method | Endpoint | Description |
| --- | --- | --- |
| **GET** | `/movies` | Fetch all available trending movies |
| **GET** | `/projections` | Fetch all scheduled movie showtimes |
| **GET** | `/projections/:id` | Fetch specific showtimes filtered by Movie ID |
| **GET** | `/seats/:id` | Get interactive occupancy seat map for a Projection ID |
| **POST** | `/book` | Submit a request to reserve a specific seat |
| **GET** | `/recommendations` | Fetch movie recommendations (Requires `?user_id=XYZ`) |

To make a single booking request:

```bash

curl -s -X POST http://localhost:8080/book \
  -H "Content-Type: application/json" \
  -d '{
    "projection_id":"p101",
    "seat_id":1,
    "user_id":"test-user"
  }'

```

## 🧪 Testing & Resilience Scenarios

You can perform the following stress tests and experiments locally to verify the system's fault tolerance and consistency. 

### 1. Circuit Breaker Evaluation

Test how the API Gateway protects the system when the **Booking Service** is down.The booking flow is protected by a circuit breaker. If the Booking Service experiences 3 or more consecutive internal errors or becomes unreachable, the Gateway will fail-fast and immediately return a `503 Service Unavailable` error code to protect database integrity:

> *"The booking system is temporarily overloaded. Please try again in a few seconds."*

* **Scenario:** Intentional service failure.
* **Action:** Stop the booking container and attempt multiple requests.

```bash
# Stop one booking service instance
docker stop $(docker ps -q --filter "name=booking" | head -n 1)

# Stress test
for i in {1..5}; do
  curl -s -X POST http://localhost:8080/book \
  -H "Content-Type: application/json" \
  -d '{
    "projection_id":"p101",
    "seat_id":1,
    "user_id":"test-user"
  }'

  echo -e "\n"
done

```

* **Expected Result:** After 3 failures (default threshold), the Circuit Breaker trips. Subsequent requests will return `503 Service Unavailable` immediately without attempting to contact the service.

---

### 2. Concurrency & Race Condition Test

Verify that the **Distributed Redis Lock** prevents two users from booking the same seat at the same exact time.

* **Scenario:** 20 users competing for seat #25.

```bash
for i in {1..20}; do
(
  result=$(curl -s -X POST http://localhost:8080/book \
    -H "Content-Type: application/json" \
    -d "{
      \"projection_id\":\"p201\",
      \"seat_id\":25,
      \"user_id\":\"user$i\"
    }" | jq -r '.message')

  echo "user$i -> $result"
) &
done

wait

```

* **Expected Result:** Only **one** request succeeds with a "Booking Successful" message. The other 19 requests will return an error (e.g., "Seat already occupied" or "Lock acquisition failed").You can try with multiple booking istances too.

---

### 3. Asynchronous Recovery (Notification Service)

Verify that **NATS JetStream** prevents message loss if the Notification Service is offline.

* **Scenario:** Delayed event processing.

```bash
# Stop notification service
docker stop $(docker ps -q --filter "name=notification" | head -n 1)

# Book a seat
curl -X POST http://localhost:8080/book \
-H "Content-Type: application/json" \
-d '{
  "projection_id":"p201",
  "seat_id":10,
  "user_id":"event-test"
}'

# Restart notification service
docker start $(docker ps -aq --filter "name=notification" | head -n 1)

# Check logs
docker logs -f $(docker ps -q --filter "name=notification" | head -n 1)
```

* **Expected Result:** As soon as the service restarts, it will automatically pull the "pending" message from NATS and process the notification. Check the logs with `docker logs -f notification-service`.

---

### 4. High Availability & Load Balancing

Test the system's ability to stay online while individual instances are killed.

* **Scenario:** Scaling and Service Replication.

```bash
# 1. Scale the booking service to 3 instances
# This creates 3 replica containers of the booking service
docker compose up -d --scale booking=3

# 2. Monitor logs to see Round-Robin in action
# Observe how the gateway distributes gRPC requests among the 3 instances
docker logs -f $(docker ps -q --filter "name=gateway" | head -n 1)

# 3. In another terminal, kill one of the booking instances
# This command finds the ID of one running booking container and stops it
# Stop one booking instance
docker stop $(docker ps -q --filter "name=booking" | head -n 1)

```

* **Expected Result:** The API Gateway (via gRPC load balancing) will detect the failure and reroute all traffic to the remaining 2 healthy instances. No downtime should be experienced by the user.

---

### 🧹 Cleanup Tests

To reset the database and event streams after testing:

```bash
docker compose down -v
docker compose up --build

```

