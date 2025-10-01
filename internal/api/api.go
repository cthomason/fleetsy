package api

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	// import the interface
	"fleetsy/pkg/api"
)

// Server implements the generated ServerInterface.
type Server struct {
	deviceMutex        sync.RWMutex
	deviceHeartbeatMap map[string][]time.Time
	deviceStatsMap     map[string][]DeviceStats
}

// struct for the device stats array
type DeviceStats struct {
	SentAt     time.Time `json:"sent_at"`
	UploadTime int64     `json:"upload_time"` // upload time is in nanoseconds
}

// struct for the incoming heartbeat POST requests
type HeartbeatPost struct {
	SentAt string `json:"sent_at"`
}

// struct for the incoming stats POST requests
type StatsPost struct {
	SentAt     string `json:"sent_at"`
	UploadTime int64  `json:"upload_time"` // upload time is in nanoseconds, use int64
}

// response struct for the stats GET requests
type StatsGet struct {
	Uptime        float32 `json:"uptime"`
	AvgUploadTime string  `json:"avg_upload_time"`
}

// NewServer creates a new instance with the required dependencies
func NewServer(deviceDB map[string][]time.Time, statsDB map[string][]DeviceStats) *Server {
	return &Server{
		deviceHeartbeatMap: deviceDB,
		deviceStatsMap:     statsDB,
	}
}

// Ensure that Server implements the ServerInterface at compile time.
var _ api.ServerInterface = (*Server)(nil)

// (POST /devices/{device_id}/heartbeat)
func (s *Server) PostDevicesDeviceIdHeartbeat(w http.ResponseWriter, r *http.Request, deviceId string) {
	// read the new heartbeat
	var newData HeartbeatPost
	if err := json.NewDecoder(r.Body).Decode(&newData); err != nil {
		// Normally I'd use http.StatusBadRequest but the spec says the only http codes allowed here are 204, 404, and 500
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	}

	// lock the mutex for writing
	s.deviceMutex.Lock()
	// defer to guarantee it's unlocked later
	defer s.deviceMutex.Unlock()
	// validate the device exists in the db
	_, found := s.deviceHeartbeatMap[deviceId]
	// return 404 if not found
	if !found {
		errorResponse := api.Error{Code: http.StatusNotFound, Message: "Device not found"}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	// parse the timestamp
	newTimestamp, tsError := time.Parse(time.RFC3339, newData.SentAt)
	if tsError != nil {
		// Normally I'd use http.StatusBadRequest but the spec says the only http codes allowed here are 204, 404, and 500
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	}

	// insert the new data
	s.deviceHeartbeatMap[deviceId] = append(s.deviceHeartbeatMap[deviceId], newTimestamp)

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
	deviceHeartbeats, heartbeatFound := s.deviceHeartbeatMap[deviceId]
	deviceStats, statsFound := s.deviceStatsMap[deviceId]
	// return 404 if not found
	if !heartbeatFound || !statsFound {
		errorResponse := api.Error{Code: http.StatusNotFound, Message: "Device not found"}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	// calculate uptime
	var uptime float32
	// check array length first
	if len(deviceHeartbeats) == 0 {
		uptime = 0.0
	} else {
		sumHeartbeats := len(deviceHeartbeats)
		firstTimestamp := deviceHeartbeats[0]
		// assuming that all heartbeats were received in chronological order
		lastTimestamp := deviceHeartbeats[len(deviceHeartbeats)-1]
		// the devices are expected to send one heartbeat every minute
		// so we need to use the number of minutes to calculate the uptime properly
		// subtract the timestamps and convert
		diff := lastTimestamp.Sub(firstTimestamp).Minutes()
		// now calculate the uptime percentage
		uptime = (float32(sumHeartbeats) / float32(diff)) * 100
	}

	// calculate upload time
	var uploadTime string = ""
	// check array length first
	if len(deviceStats) == 0 {
		uploadTime = ""
	} else {
		var totalSeconds float64 = 0
		for _, stats := range deviceStats {
			// convert to time.Duration to make some of this easier
			dur := time.Duration(stats.UploadTime)
			totalSeconds += dur.Seconds()
		}
		avg := totalSeconds / float64(len(deviceStats))
		// time.Duration works in nanoseconds, so we need to convert seconds as part of this
		// there are 1e9 nanoseconds in every second
		uploadTime = time.Duration(avg * 1e9).String()
	}
	response := StatsGet{
		Uptime:        uptime,
		AvgUploadTime: uploadTime,
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
		// Normally I'd use http.StatusBadRequest but the spec says the only http codes allowed here are 204, 404, and 500
		http.Error(w, "Invalid request body", http.StatusInternalServerError)
		return
	}

	// lock the mutex for writing
	s.deviceMutex.Lock()
	// defer to guarantee it's unlocked later
	defer s.deviceMutex.Unlock()
	// validate the device exists in the db
	_, found := s.deviceStatsMap[deviceId]
	// return 404 if not found
	if !found {
		errorResponse := api.Error{Code: http.StatusNotFound, Message: "Device not found"}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	// parse the timestamp
	newTimestamp, tsError := time.Parse(time.RFC3339, newData.SentAt)
	if tsError != nil {
		// Normally I'd use http.StatusBadRequest but the spec says the only http codes allowed here are 204, 404, and 500
		http.Error(w, "Invalid request body", http.StatusInternalServerError)
		return
	}

	newDeviceStats := DeviceStats{
		SentAt:     newTimestamp,
		UploadTime: newData.UploadTime,
	}

	s.deviceStatsMap[deviceId] = append(s.deviceStatsMap[deviceId], newDeviceStats)

	// send conformation of success
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}
