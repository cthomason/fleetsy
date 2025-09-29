package api

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	// import the interface
	"fleetsy/pkg/api"
)

// Server implements the generated ServerInterface.
type Server struct {
	deviceMutex sync.RWMutex
	deviceMap   map[string][]time.Time
}

type HeartbeatPost struct {
	SentAt string `json:"sent_at"`
}

type StatsPost struct {
	SentAt     string `json:"sent_at"`
	UploadTime int    `json:"upload_time"`
}

type StatsGet struct {
	Uptime        int    `json:"uptime"`
	AvgUploadTime string `json:"avg_upload_time"`
}

// NewServer creates a new instance with the required dependencies
func NewServer(deviceDB map[string][]time.Time) *Server {
	return &Server{
		deviceMap: deviceDB,
	}
}

// Ensure that Server implements the ServerInterface at compile time.
var _ api.ServerInterface = (*Server)(nil)

// (POST /devices/{device_id}/heartbeat)
func (s *Server) PostDevicesDeviceIdHeartbeat(w http.ResponseWriter, r *http.Request, deviceId string) {
	// read the new heartbeat
	var newData HeartbeatPost
	if err := json.NewDecoder(r.Body).Decode(&newData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
	}

	// lock the mutex for writing
	s.deviceMutex.Lock()
	// defer to guarantee it's unlocked later
	defer s.deviceMutex.Unlock()
	// validate the device exists in the db
	_, found := s.deviceMap[deviceId]
	// return 404 if not found
	if !found {
		errorResponse := api.Error{Code: 404, Message: "Device not found"}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	// parse the timestamp
	newTimestamp, tsError := time.Parse(time.RFC3339, newData.SentAt)
	if tsError != nil {
		errorResponse := api.Error{Code: 501, Message: "Invalid timestamp"}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}
	log.Printf("newTimestamp is %s", newTimestamp)

	// insert the new data
	s.deviceMap[deviceId] = append(s.deviceMap[deviceId], newTimestamp)
	log.Printf("deviceMap is %s\n", s.deviceMap)

	// send conformation of success
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}

// (GET /devices/{device_id}/stats)
func (s *Server) GetDevicesDeviceIdStats(w http.ResponseWriter, r *http.Request, deviceId string) {

	// lock the mutex for reading
	s.deviceMutex.Lock()
	// defer to guarantee it's unlocked later
	defer s.deviceMutex.Unlock()
	// validate the device exists in the db
	deviceStats, found := s.deviceMap[deviceId]
	// return 404 if not found
	if !found {
		errorResponse := api.Error{Code: 404, Message: "Device not found"}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	log.Printf("device stats %s", deviceStats)
	response := StatsGet{
		Uptime:        7,
		AvgUploadTime: "72",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// (POST /devices/{device_id}/stats)
func (s *Server) PostDevicesDeviceIdStats(w http.ResponseWriter, r *http.Request, deviceId string) {
	// read the new heartbeat
	var newData StatsPost
	if err := json.NewDecoder(r.Body).Decode(&newData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
	}

	log.Printf("newData is %s", newData.SentAt)

	// lock the mutex for writing
	s.deviceMutex.Lock()
	// defer to guarantee it's unlocked later
	defer s.deviceMutex.Unlock()
	// validate the device exists in the db
	_, found := s.deviceMap[deviceId]
	// return 404 if not found
	if !found {
		errorResponse := api.Error{Code: 404, Message: "Device not found"}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	// parse the timestamp
	newTimestamp, tsError := time.Parse(time.RFC3339, newData.SentAt)
	if tsError != nil {
		errorResponse := api.Error{Code: 501, Message: "Invalid timestamp"}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}
	log.Printf("newTimestamp is %s", newTimestamp)

	s.deviceMap[deviceId] = append(s.deviceMap[deviceId], newTimestamp)
	log.Printf("deviceMap is %s\n", s.deviceMap)

	// send conformation of success
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}
