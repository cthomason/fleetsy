package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"

	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	// Your local packages
	handlers "fleetsy/internal/api"
	api "fleetsy/pkg/api"
)

func main() {

	// open the fleetsy file
	file, err := os.Open("devices.csv")

	if err != nil {
		log.Fatal("Failed to read devices.csv", err)
	}

	// read all the records, including the column headers
	// this is a small proof of concept so we can get away with loading everything into memory
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()

	if err != nil {
		fmt.Println("Error reading records")
	}

	// set up the data structures to hold the incoming data
	for _, eachrecord := range records {
		fmt.Println(eachrecord)
	}

	// Clean up the file because we don't need it anymore
	file.Close()

	// Initialize api server
	apiServer := &handlers.Server{}

	router := chi.NewRouter()

	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	// register the api handlers
	// api.RegisterHandlers(router, apiServer)
	handler := api.HandlerFromMux(apiServer, router)

	// start the server
	port := 8080
	log.Printf("Server is running on port %d\n", port)
	httpErr := http.ListenAndServe(fmt.Sprintf(":%d", port), handler)
	if httpErr != nil {
		log.Fatalf("Failed to start server: %v", httpErr)
	}
}
