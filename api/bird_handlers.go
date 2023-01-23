package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Bird struct {
	Species     string `json:"species"`
	Description string `json:"description"`
}

var birds []Bird

func getBirdHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("get")
	// Convert the birds struct to a json
	birdListBytes, err := json.Marshal(birds)

	if err != nil {
		fmt.Println(fmt.Errorf("Error: %v", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(birdListBytes)
}

func createBirdHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("create API")
	bird := Bird{}
	w.Header().Set("Content.Type", "application/json")
	_ = json.NewDecoder(r.Body).Decode(&bird)

	birds = append(birds, bird)
	// Redoing the get
	json.NewEncoder(w).Encode(bird)
}
