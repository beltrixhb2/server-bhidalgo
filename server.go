package main

import (
	"fmt"
	"encoding/json"
	"net/http"
	"time"
	loggly "github.com/jamespearly/loggly"
	"github.com/gorilla/mux"

)

type StatusResponse struct {
	Time   string `json:"time"`
	Status int    `json:"status"`
}

type statusResponseWriter struct {
    http.ResponseWriter
    statusCode int
}

func NewStatusResponseWriter(w http.ResponseWriter) *statusResponseWriter {
    return &statusResponseWriter{
        ResponseWriter: w,
        statusCode:     http.StatusOK,
    }
}

func (sw *statusResponseWriter) WriteHeader(statusCode int) {
    sw.statusCode = statusCode
    sw.ResponseWriter.WriteHeader(statusCode)
}

func main() {	
	client := loggly.New("Server")
	r := mux.NewRouter()
	r.Use(RequestLoggerMiddleware(r,client))
	r.HandleFunc("/bhidalgo/status", statusHandler).Methods(http.MethodGet)
	r.PathPrefix("/").HandlerFunc(notFoundHandler).Methods(http.MethodGet)
	r.PathPrefix("/").HandlerFunc(notAllowedHandler)
	r.NotFoundHandler = http.HandlerFunc(notFoundHandler)
	fmt.Printf("Server is running on :39000\n")
	http.ListenAndServe("0.0.0.0:39000",r) 
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	response := StatusResponse{
		Time:   time.Now().Format(time.RFC3339),
		Status: http.StatusOK,
	}

	sendJSONResponse(w, response)
}

func notAllowedHandler(w http.ResponseWriter, r *http.Request) {
    http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r)
}

func sendJSONResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}

func RequestLoggerMiddleware(r *mux.Router, client *loggly.ClientType) mux.MiddlewareFunc {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		sw := NewStatusResponseWriter(w)
		defer func(){
			logMessage := fmt.Sprintf("Method: %s, IP: %s, Path: %s, Status: %d",
                        	req.Method, req.RemoteAddr, req.URL.Path, sw.statusCode)
                	client.EchoSend("info",logMessage)
		}()
		fmt.Println("Hello")
            	next.ServeHTTP(sw, req)
		
        })
    }
}

