package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"time"

	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	// Your local packages
	handlers "fleetsy/internal/api"
	api "fleetsy/pkg/api"
)

func main() {

	// open the devices file
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
	deviceHeartbeatMap := make(map[string][]time.Time)
	deviceStatsMap := make(map[string][]handlers.DeviceStats)
	// load the device names and initialize an empty array
	for _, eachrecord := range records {
		deviceHeartbeatMap[eachrecord[0]] = []time.Time{}
		deviceStatsMap[eachrecord[0]] = []handlers.DeviceStats{}
	}

	// Clean up the file because we don't need it anymore
	file.Close()

	// Initialize api server
	apiServer := handlers.NewServer(deviceHeartbeatMap, deviceStatsMap)

	// initialize api router
	apiRouter := chi.NewRouter()
	// register the handlers
	apiHandler := api.HandlerFromMux(apiServer, apiRouter)

	// create main router so we can put the handlers on the right path
	mainRouter := chi.NewRouter()
	mainRouter.Use(middleware.Logger)
	mainRouter.Use(middleware.Recoverer)

	// healthcheck endpoint
	mainRouter.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})

	// load the api handlers to the proper path
	mainRouter.Mount("/api/v1", apiHandler)

	// for debugging purposes
	// log.Println(("registered routes"))
	// walkFunc := func(method string, route string, handler http.Handler, millewares ...func(http.Handler) http.Handler) error {
	// 	route = strings.Replace(route, "/*/", "/", -1)
	// 	log.Printf("%s %s\n", method, route)
	// 	return nil
	// }
	// if err := chi.Walk(mainRouter, walkFunc); err != nil {
	// 	log.Panicf("Error walking routes: %s\n", err.Error())
	// }

	// start the server
	port := 8080
	log.Printf("Server is running on port %d\n", port)
	httpErr := http.ListenAndServe(fmt.Sprintf(":%d", port), mainRouter)
	if httpErr != nil {
		log.Fatalf("Failed to start server: %v", httpErr)
	}
}
