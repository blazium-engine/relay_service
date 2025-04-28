package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// Constants
const BASE_FORWARD_URL = "https://api.pogr.io/v1/intake"

func initHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read headers
	pogrClient := r.Header.Get("POGR_CLIENT")
	pogrBuild := r.Header.Get("POGR_BUILD")

	if pogrClient == "" || pogrBuild == "" {
		http.Error(w, "Missing required headers", http.StatusBadRequest)
		return
	}

	// Read the original body
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	// Create a new request to forward
	forwardURL := BASE_FORWARD_URL + "/init"
	req, err := http.NewRequest(http.MethodPost, forwardURL, bytes.NewReader(bodyBytes))
	if err != nil {
		http.Error(w, "Failed to create forward request", http.StatusInternalServerError)
		return
	}

	// Copy relevant headers
	req.Header.Set("Content-Type", r.Header.Get("Content-Type"))
	req.Header.Set("POGR_CLIENT", pogrClient)
	req.Header.Set("POGR_BUILD", pogrBuild)

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to forward request", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy the response from the forwarded request
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func genericHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read required header
	intakeSessionID := r.Header.Get("INTAKE_SESSION_ID")
	if intakeSessionID == "" {
		http.Error(w, "Missing INTAKE_SESSION_ID header", http.StatusBadRequest)
		return
	}

	// Forward URL is based on the incoming path
	forwardURL := BASE_FORWARD_URL + r.URL.Path

	// Create the forwarding request
	req, err := http.NewRequest(http.MethodPost, forwardURL, r.Body)
	if err != nil {
		http.Error(w, "Failed to create forward request", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	// Set headers
	req.Header.Set("Content-Type", r.Header.Get("Content-Type"))
	req.Header.Set("INTAKE_SESSION_ID", intakeSessionID)

	// Set Content-Length if available
	if r.ContentLength > 0 {
		req.ContentLength = r.ContentLength
	}

	// Send the request
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to forward request", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Forward response status code
	w.WriteHeader(resp.StatusCode)

	// Stream response body back to client
	io.Copy(w, resp.Body)
}


func main() {
	http.HandleFunc("/init", initHandler)
	http.HandleFunc("/data", genericHandler)
	http.HandleFunc("/event", genericHandler)
	http.HandleFunc("/logs", genericHandler)
	http.HandleFunc("/metrics", genericHandler)
	http.HandleFunc("/monitor", genericHandler)
	http.HandleFunc("/end", genericHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server running on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
