package api

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"sync"

	// import the interface
	"fleetsy/pkg/api"
)

// Server implements the generated ServerInterface.
type Server struct {
	deviceMutex sync.RWMutex
	deviceMap   map[string]string
}

// NewServer creates a new instance with the required dependencies
func NewServer() *Server {
	return &Server{
		deviceMap: make(map[string]string),
	}
}

type PostHeartbeat struct {
	SentAt string `json:"sent_at"`
}

type PostStats struct {
	SentAt     string `json:"sent_at"`
	UploadTime int    `json:"upload_time"`
}

type GetStats struct {
	Uptime        int    `json:"uptime"`
	AvgUploadTime string `json:"avg_upload_time"`
}

// Ensure that Server implements the ServerInterface at compile time.
var _ api.ServerInterface = (*Server)(nil)

// (POST /devices/{device_id}/heartbeat)
func (s *Server) PostDevicesDeviceIdHeartbeat(w http.ResponseWriter, r *http.Request, deviceId string) {
	// read the new heartbeat
	var newData PostHeartbeat
	if err := json.NewDecoder(r.Body).Decode(&newData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
	}

	// for debugging purposes
	log.Printf("newData is %s", newData.SentAt)

	// lock the mutext for writing
	s.deviceMutex.Lock()
	// defer to guarantee it's unlocked later
	defer s.deviceMutex.Unlock()
	// validate the device exists in the db
	_, found := s.deviceMap[deviceId]
	if !found {
		errorResponse := api.Error{Code: 404, Message: "Device not found"}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}
	s.deviceMap[deviceId] = newData.SentAt

	// send conformation of success
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}

// (GET /devices/{device_id}/stats)
func (s *Server) GetDevicesDeviceIdStats(w http.ResponseWriter, r *http.Request, deviceId string) {

	response := GetStats{
		Uptime:        7,
		AvgUploadTime: "72",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	// sent if device not found
	// errorResponse := api.Error{
	// 	Code:    404,
	// 	Message: "Device not found",
	// }
	// w.Header().Set("Content-Type", "application/json")
	// w.WriteHeader(http.StatusNotFound)
	// json.NewEncoder(w).Encode(errorResponse)
}

// (POST /devices/{device_id}/stats)
func (s *Server) PostDevicesDeviceIdStats(w http.ResponseWriter, r *http.Request, deviceId string) {
	// read the new heartbeat
	var newData PostStats
	if err := json.NewDecoder(r.Body).Decode(&newData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
	}

	log.Printf("newData is %s", newData.SentAt)

	// lock the mutext for writing
	s.deviceMutex.Lock()
	// defer to guarantee it's unlocked later
	defer s.deviceMutex.Unlock()
	// validate the device exists in the db
	_, found := s.deviceMap[deviceId]
	if !found {
		errorResponse := api.Error{Code: 404, Message: "Device not found"}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}
	s.deviceMap[deviceId] = newData.SentAt

	// send conformation of success
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)

	var requestBody []byte
	requestBody, _ = io.ReadAll(r.Body)
	r.Body = io.NopCloser(bytes.NewBuffer(requestBody))
	log.Printf("body is %s", string(requestBody))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}
