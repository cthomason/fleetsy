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
	deviceMutex        sync.RWMutex
	deviceHeartbeatMap map[string][]time.Time
	deviceStatsMap     map[string][]DeviceStats
}

type DeviceStats struct {
	SentAt     time.Time `json:"sent_at"`
	UploadTime int64     `json:"upload_time"`
}

type HeartbeatPost struct {
	SentAt string `json:"sent_at"`
}

type StatsPost struct {
	SentAt     string `json:"sent_at"`
	UploadTime int64  `json:"upload_time"` // upload time is in nanoseconds, use int64
}

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
		http.Error(w, "Invalid request body", http.StatusBadRequest)
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
		errorResponse := api.Error{Code: http.StatusBadRequest, Message: "Invalid timestamp"}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorResponse)
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
		log.Printf("Calculating uptime")
		sumHeartbeats := len(deviceHeartbeats)
		firstTimestamp := deviceHeartbeats[0]
		log.Printf("First timestamp %s", firstTimestamp.String())
		lastTimestamp := deviceHeartbeats[len(deviceHeartbeats)-1]
		log.Printf("lastTimestamp %s", lastTimestamp.String())
		diff := lastTimestamp.Sub(firstTimestamp).Minutes()
		// log.Printf("diff is %s", diff.String())
		uptime = (float32(sumHeartbeats) / float32(diff)) * 100
	}
	// calculate upload time
	var uploadTime string = ""
	// check array length first
	if len(deviceStats) == 0 {
		log.Println("deviceStatns length is zero")
		uploadTime = ""
	} else {
		log.Println("calculating average")
		var totalSeconds float32 = 0
		for _, stats := range deviceStats {
			// need to convert uploadTime from nanoseconds to seconds
			uploadTimeSeconds := float32(stats.UploadTime / 1e9)
			totalSeconds += float32(uploadTimeSeconds)
		}
		log.Printf("total is %f\n", totalSeconds)
		log.Printf("len is %d\n", len(deviceStats))
		avg := totalSeconds / float32(len(deviceStats))
		log.Printf("avg %f\n", avg)
		// convert to string
		uploadTime = time.Duration(avg * 1e9).String()
	}
	log.Printf("uploadTime is %s", uploadTime)
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
		http.Error(w, "Invalid request body", http.StatusBadRequest)
	}

	log.Printf("newData is %s", newData.SentAt)

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
		errorResponse := api.Error{Code: http.StatusBadRequest, Message: "Invalid timestamp"}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorResponse)
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
