package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	// import the interface
	"fleetsy/pkg/api"
)

// Server implements the generated ServerInterface.
type Server struct{}

// Ensure that Server implements the ServerInterface at compile time.
var _ api.ServerInterface = (*Server)(nil)

func (s *Server) PostDevicesDeviceIdHeartbeat(w http.ResponseWriter, r *http.Request, deviceId string) {
	fmt.Println("Received request for device %s", deviceId)

	// If the pet is not found, return a 404 error.
	errorResponse := api.Error{
		Code:    404,
		Message: "Device not found",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(errorResponse)
}

func (s *Server) GetDevicesDeviceIdStats(w http.ResponseWriter, r *http.Request, deviceId string) {
	fmt.Println("Received request for device %s", deviceId)

	errorResponse := api.Error{
		Code:    404,
		Message: "Device not found",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(errorResponse)
}

func (s *Server) PostDevicesDeviceIdStats(w http.ResponseWriter, r *http.Request, deviceId string) {
	fmt.Println("Received request for device %s", deviceId)

	errorResponse := api.Error{
		Code:    404,
		Message: "Device not found",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(errorResponse)
}
