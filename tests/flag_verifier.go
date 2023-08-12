package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
)

type ResponseItem struct {
	Flag    string
	Message string
}

func flagStatus() string {
	randNum := rand.Intn(100)
	switch {
	case randNum < 70:
		return "accepted"
	case randNum < 80:
		return "invalid"
	case randNum < 90:
		return "too old"
	default:
		return "is not available"
	}
}
func verifyFlagsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	var flags []string
	err := json.NewDecoder(r.Body).Decode(&flags)
	if err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}
	var response []ResponseItem
	for _, item := range flags {
		response = append(response, ResponseItem{Flag: item, Message: flagStatus()})
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Error creating response payload", http.StatusInternalServerError)
		return
	}
	fmt.Println("Flags", response, "verified")
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)

}

func main() {
	http.HandleFunc("/", verifyFlagsHandler)

	fmt.Println("Server listening on port 8000...")
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		fmt.Println("Error starting server: ", err)
	}
}
