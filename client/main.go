package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// ==============================
//  CONFIG
// ==============================

const baseURL = "http://localhost:8080"

// ==============================
//  MODELS
// ==============================

type Movie struct {
	Id    string `json:"id"`
	Title string `json:"title"`
}

type MovieList struct {
	Movies []Movie `json:"movies"`
}

type Projection struct {
	Id         string `json:"id"`
	MovieId    string `json:"movie_id"`
	ShowTime   string `json:"show_time"`
	RoomId     string `json:"room_id"`
	TotalSeats int    `json:"total_seats"`
}

type ProjectionList struct {
	Projections []Projection `json:"projections"`
}

type SeatsResponse struct {
	ProjectionID string `json:"projection_id"`
	Seats        []struct {
		SeatID   int  `json:"seat_id"`
		Occupied bool `json:"occupied"`
	} `json:"seats"`
}

type RecommendationResponse struct {
	Movies []Movie `json:"movies"`
}

// ==============================
//  UI HELPERS 
// ==============================

func header() {
	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("🎬  CINEMA STREAMING PLATFORM")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

func section(title string) {
	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("📺 %s\n", strings.ToUpper(title))
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━")
}

func card(title string, subtitle string) {
	fmt.Printf("🎬 %s\n", title)
	fmt.Printf("   %s\n", subtitle)
	fmt.Println("──────────────────────")
}

func input(prompt string) string {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	return strings.TrimSpace(text)
}

// ==============================
//  MAIN
// ==============================

func main() {
	reader := bufio.NewReader(os.Stdin)

	for {
		header()

		fmt.Println("1️⃣  Browse Movies")
		fmt.Println("2️⃣  View Projections")
		fmt.Println("3️⃣  View Seats")
		fmt.Println("4️⃣  Book Seat")
		fmt.Println("5️⃣  My Recommendations")
		fmt.Println("0️⃣  Exit")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Print("👉 Choose: ")

		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {

		case "1":
			getMovies()

		case "2":
			movieID := input("🎬 Enter Movie ID: ")
			getProjections(movieID)

		case "3":
			id := input("🎟️ Enter Projection ID: ")
			getSeats(id)

		case "4":
			handleBooking(reader)

		case "5":
			user := input("👤 Enter User ID: ")
			getRecommendations(user)

		case "0":
			fmt.Println("👋 See you next time!")
			return

		default:
			fmt.Println("❌ Invalid choice")
		}

		fmt.Println("\n🔁 Press ENTER to continue...")
		reader.ReadString('\n')
	}
}

// ==============================
//  MOVIES
// ==============================

func getMovies() {
	section("Trending Movies")

	resp, err := http.Get(baseURL + "/movies")
	if err != nil {
		fmt.Println("❌ Error:", err)
		return
	}
	defer resp.Body.Close()

	var data MovieList
	json.NewDecoder(resp.Body).Decode(&data)

	for _, m := range data.Movies {
		card(m.Title, "ID: "+m.Id)
	}
}

// ==============================
//  PROJECTIONS
// ==============================

func getProjections(movieID string) {
	if movieID == "" {
		fmt.Println("❌ Invalid movie_id")
		return
	}

	section("Available Showtimes")

	resp, err := http.Get(baseURL + "/projections/" + movieID)
	if err != nil {
		fmt.Println("❌ Error:", err)
		return
	}
	defer resp.Body.Close()

	var data ProjectionList
	json.NewDecoder(resp.Body).Decode(&data)

	if len(data.Projections) == 0 {
		fmt.Println("🚫 No projections found")
		return
	}

	for _, p := range data.Projections {
		card(
			"Projection "+p.Id,
			fmt.Sprintf("⏰ %s | 🎭 Room: %s | 💺 Seats: %d",
				p.ShowTime, p.RoomId, p.TotalSeats),
		)
	}
}

// ==============================
//  SEATS
// ==============================

func getSeats(projectionID string) {
	if projectionID == "" {
		fmt.Println("❌ Invalid projection_id")
		return
	}

	section("Seat Map")

	resp, err := http.Get(baseURL + "/seats/" + projectionID)
	if err != nil {
		fmt.Println("❌ Error:", err)
		return
	}
	defer resp.Body.Close()

	var data SeatsResponse
	json.NewDecoder(resp.Body).Decode(&data)
	
	// CONTROLLO CRITICO: Se non ci sono posti, la proiezione non esiste o è vuota
	if len(data.Seats) == 0 {
		fmt.Printf("🚫 No data found for Projection ID: %s. Please check the ID and try again.\n", projectionID)
		return
	}

	fmt.Printf("🎟️ Projection: %s\n\n", data.ProjectionID)

	for _, s := range data.Seats {
		status := "🟩 Available"
		if s.Occupied {
			status = "🟥 Occupied"
		}
		fmt.Printf("Seat %-3d %s\n", s.SeatID, status)
	}
}

// ==============================
// BOOKING
// ==============================

func handleBooking(reader *bufio.Reader) {

	fmt.Println("\n🎟️ BOOK YOUR SEAT")

	proj := input("Projection ID: ")
	seatStr := input("Seat ID: ")
	user := input("User ID: ")

	seatID, err := strconv.Atoi(seatStr)
	if err != nil {
		fmt.Println("❌ Invalid seat")
		return
	}

	payload := map[string]interface{}{
		"projection_id": proj,
		"seat_id":       seatID,
		"user_id":       user,
	}

	body, _ := json.Marshal(payload)

	resp, err := http.Post(baseURL+"/book", "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Println("❌ Error:", err)
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	section("Booking Result")

	for k, v := range result {
		fmt.Printf(" %s: %v\n", strings.Title(k), v)
	}
}

// ==============================
// RECOMMENDATIONS
// ==============================

func getRecommendations(userID string) {
	if userID == "" {
		fmt.Println("❌ Invalid user_id")
		return
	}

	section("Recommended For You")

	resp, err := http.Get(baseURL + "/recommendations?user_id=" + userID)
	if err != nil {
		fmt.Println("❌ Error:", err)
		return
	}
	defer resp.Body.Close()

	var data RecommendationResponse
	json.NewDecoder(resp.Body).Decode(&data)

	for i, m := range data.Movies {
		card(
			fmt.Sprintf("%d. %s", i+1, m.Title),
			"Movie ID: "+m.Id,
		)
	}
}
