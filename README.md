
# 🎬 CineFlix: Microservices Cinema Reservation System

Welcome to **CineFlix**, a high-performance, event-driven cinema reservation platform built using Go, gRPC, and the Gin Web Framework. 

This project features an interactive Netflix-style CLI Client, an API Gateway with built-in resilience, and multiple distributed backend microservices communicating via gRPC and event streams.

---

## 🌐 Live Demo & Hosting Notice

The production backend of this architecture is currently deployed and hosted on an **AWS EC2 Instance**:
* **API Gateway URL:** `http://3.232.139.1:8080`

You can run the local CLI client immediately to interact with this live cloud deployment without setting up the backend cluster on your machine.

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

### 🏎️ Option 1: Quick Start (Connect to Live AWS EC2 Backend)

If you just want to test the application immediately using our hosted infrastructure, follow these steps:

1. Locate your standalone CLI `main.go` file.
2. Ensure the top configuration constant points to the live server:
```go
   const baseURL = "[http://3.232.139.1:8080](http://3.232.139.1:8080)"

```

3. Run the application from your terminal:

```bash
   go run main.go

```

---

### 🐳 Option 2: Run the Entire System Locally

To spin up the complete isolated microservices environment on your local machine using Docker Compose:

#### 1. Update the Client Configuration

Open the CLI client file (`main.go`) and change the `baseURL` to target your local machine:

```go
const baseURL = "http://localhost:8080"

```

#### 2. Start the Docker Containers

From the root directory containing your `docker-compose.yml`, run:

```bash
docker compose up --build

```
Launch & Scale Infrastructure: Open a terminal in the root directory of the project. To test how the API Gateway balances traffic across multiple instances of a service using the internal round-robin gRPC balancer, you can scale the backend nodes dynamically:

Bash

   docker compose up --build --scale catalog-service=3 --scale booking=2

This command builds and runs the following environment:

* **nats-broker** (Ports `4222`, `8222`)
* **redis** (Port `6379`)
* **gateway** (Port `8080`)
* **catalog-service**, **booking**, **notification-service**, **recommendation-service**

#### 3. Run the CLI Client

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

### Raw cURL Interaction Examples

**Get Seating Map:**

```bash
curl -X GET [http://3.232.139.1:8080/seats/proj-1002](http://3.232.139.1:8080/seats/proj-1002)

```

**Book a Seat:**

```bash
curl -X POST [http://3.232.139.1:8080/book](http://3.232.139.1:8080/book) \
     -H "Content-Type: application/json" \
     -d '{"projection_id": "proj-1002", "seat_id": 42, "user_id": "user-abc"}'

```


### Circuit Breaker Triggers

The booking flow is protected by a circuit breaker. If the Booking Service experiences 3 or more consecutive internal errors or becomes unreachable, the Gateway will fail-fast and immediately return a `503 Service Unavailable` error code to protect database integrity:

> *"The booking system is temporarily overloaded. Please try again in a few seconds."*

```

```
