package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
)

var apiHost = os.Getenv("API_HOST")
var apiPort = os.Getenv("API_PORT")

func index(w http.ResponseWriter, r *http.Request) {
	tpl, _ := template.ParseFiles("./assets/index.html")
	data := map[string]string{
		"API_HOST": apiHost,
		"API_PORT": apiPort,
	}
	w.WriteHeader(http.StatusOK)
	tpl.Execute(w, data)
}

func main() {
	// fs := http.FileServer(http.Dir("./assets"))
	http.HandleFunc("/", index)
	fmt.Println("Starting server in port 8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
