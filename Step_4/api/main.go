package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/rs/cors"
)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

var (
	host     = getEnv("DB_HOST", "localhost")
	port     = 5455
	user     = getEnv("DB_USER", "postgresUser")
	password = getEnv("DB_PASSWORD", "postgresPW")
	dbname   = "birds_db"
)

func newRouter() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/", handler).Methods("GET")

	r.HandleFunc("/bird", getBirdHandler).Methods("GET")
	r.HandleFunc("/bird", createBirdHandler).Methods("POST")

	return r
}

func main() {

	// connString := "dbname=birds_db port=5455 user=postgresUser password=postgresPW sslmode=disable"
	connString := fmt.Sprintf("host=%s port=%d user=%s "+"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", connString)

	if err != nil {
		panic(err)
	}
	err = db.Ping()
	if err != nil {
		panic(err)
	}

	InitStore(&dbStore{db: db})

	r := newRouter()

	fmt.Println("API server in port 8090")
	// Using the cors package is a life saver for local testing
	handler := cors.Default().Handler(r)
	err = http.ListenAndServe(":8090", handler)
	if err != nil {
		panic(err.Error())
	}

}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World API!")
}
