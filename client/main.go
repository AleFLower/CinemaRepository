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

const baseURL = "http://localhost:8080"

type Movie struct {
	Id    string `json:"id"`
	Title string `json:"title"`
}

type MovieList struct {
	Movies []Movie `json:"movies"`
}

type SeatsResponse struct {
	ProjectionID string `json:"projection_id"`
	Seats []struct {
		SeatID   int  `json:"seat_id"`
		Occupied bool `json:"occupied"`
	} `json:"seats"`
}

func main() {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("\n🎬 DASHBOARD CINEMA")
		fmt.Println("1. Lista film")
		fmt.Println("2. Vedi posti")
		fmt.Println("3. Prenota posto")
		fmt.Println("0. Esci")
		fmt.Print("> ")

		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			getMovies()
		case "2":
			fmt.Print("Inserisci projection_id: ")
			id, _ := reader.ReadString('\n')
			getSeats(strings.TrimSpace(id))
		case "3":
			handleBooking(reader)
		case "0":
			return
		default:
			fmt.Println("Scelta non valida")
		}
	}
}

//
// 🎬 GET MOVIES
//
func getMovies() {
	resp, err := http.Get(baseURL + "/movies")
	if err != nil {
		fmt.Println("Errore:", err)
		return
	}
	defer resp.Body.Close()

	var data MovieList
	json.NewDecoder(resp.Body).Decode(&data)

	fmt.Println("\n📽️ Film disponibili:")
	for _, m := range data.Movies {
		fmt.Printf("- ID: %s | Titolo: %s\n", m.Id, m.Title)
	}
}

//
// 💺 GET SEATS
//
func getSeats(projectionID string) {
	resp, err := http.Get(baseURL + "/seats/" + projectionID)
	if err != nil {
		fmt.Println("Errore:", err)
		return
	}
	defer resp.Body.Close()

	var data SeatsResponse
	json.NewDecoder(resp.Body).Decode(&data)

	fmt.Printf("\n💺 Posti per proiezione %s:\n", data.ProjectionID)

	for _, s := range data.Seats {
		status := "🟢"
		if s.Occupied {
			status = "🔴"
		}
		fmt.Printf("Posto %d [%s]\n", s.SeatID, status)
	}
}

//
// 🎟️ BOOK
//
func handleBooking(reader *bufio.Reader) {
	fmt.Print("Projection ID: ")
	proj, _ := reader.ReadString('\n')

	fmt.Print("Seat ID: ")
	seatStr, _ := reader.ReadString('\n')

	fmt.Print("User ID: ")
	user, _ := reader.ReadString('\n')


	seatID, _ := strconv.Atoi(strings.TrimSpace(seatStr))

	payload := map[string]interface{}{
		"projection_id": strings.TrimSpace(proj),
		"seat_id":       seatID,
		"user_id":       strings.TrimSpace(user),
	}

	body, _ := json.Marshal(payload)

	resp, err := http.Post(baseURL+"/book", "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Println("Errore:", err)
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	fmt.Println("\n📩 Risultato prenotazione:")
	for k, v := range result {
		fmt.Printf("%s: %v\n", k, v)
	}
}
