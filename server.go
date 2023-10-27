package main

import (
	"fmt"
	"encoding/json"
	"net/http"
	"time"
	loggly "github.com/jamespearly/loggly"

)

type StatusResponse struct {
	Time   string `json:"time"`
	Status int    `json:"status"`
}

func main() {
	http.HandleFunc("/bhidalgo/status", statusHandler)
	http.HandleFunc("/", notFoundHandler)
	client := loggly.New("Server")
	fmt.Printf("Server is running on :8080")
	http.ListenAndServe(":8080", logRequest(http.DefaultServeMux, client))
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	response := StatusResponse{
		Time:   time.Now().Format(time.RFC3339),
		Status: http.StatusOK,
	}

	sendJSONResponse(w, response)
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r)
}

func sendJSONResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}

func logRequest(handler http.Handler, client *loggly.ClientType) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logMessage := fmt.Sprintf("Method: %s, IP: %s, Path: %s, Status: %d",
    			r.Method, r.RemoteAddr, r.URL.Path, http.StatusOK)
		client.EchoSend("info",logMessage) 

		handler.ServeHTTP(w, r)
	})
}

