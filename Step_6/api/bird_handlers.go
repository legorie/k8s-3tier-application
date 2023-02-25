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

	birds, err := store.GetBird()
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

	_ = json.NewDecoder(r.Body).Decode(&bird)

	// birds = append(birds, bird)
	err := store.CreateBird(&bird)
	if err != nil {
		fmt.Println(err)
	}
	// Write the inserted variable as output JSON for the API call
	json.NewEncoder(w).Encode(bird)
}
