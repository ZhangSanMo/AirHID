package server

import (
	"crypto/rand"
	"embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"airhid/internal/input"
)

//go:embed templates/index.html
var templatesFS embed.FS

type Response struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type TypeRequest struct {
	Text string `json:"text"`
	Mode string `json:"mode"`
}

type KeyRequest struct {
	Key string `json:"key"`
}

type CommandRequest struct {
	Command string `json:"command"`
}

type MouseRequest struct {
	Action string  `json:"action"`
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
}

var serverToken string

func Start(host, port, token string) error {
	serverToken = token
	
	http.HandleFunc("/", handleIndex)
	
	// Protected routes
	http.HandleFunc("/type", authMiddleware(handleType))
	http.HandleFunc("/command", authMiddleware(handleCommand))
	http.HandleFunc("/key", authMiddleware(handleKey))
	http.HandleFunc("/mouse", authMiddleware(handleMouse))
	http.HandleFunc("/api/info", authMiddleware(handleInfo))

	addr := fmt.Sprintf("%s:%s", host, port)
	return http.ListenAndServe(addr, nil)
}

func GetToken() string {
	return serverToken
}

func GenerateToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			// Fallback to query param for initial connection checks if needed, 
			// strictly speaking API calls should use header.
			// Let's enforce header for APIs.
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" || parts[1] != serverToken {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		next(w, r)
	}
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFS(templatesFS, "templates/index.html")
	if err != nil {
		http.Error(w, "Template not found: "+err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

func handleInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "online", "version": "airhid-1.1"})
}

func handleType(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req TypeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResponse(w, false, err.Error())
		return
	}

	if req.Text == "" && req.Mode == "type" {
		jsonResponse(w, false, "No text provided")
		return
	}

	if err := input.SimulateType(req.Text, req.Mode); err != nil {
		jsonResponse(w, false, err.Error())
		return
	}

	jsonResponse(w, true, "")
}

func handleMouse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req MouseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResponse(w, false, err.Error())
		return
	}

	input.SimulateMouse(req.Action, req.X, req.Y)
	jsonResponse(w, true, "")
}

func handleKey(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req KeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResponse(w, false, err.Error())
		return
	}

	if err := input.SimulateKey(req.Key); err != nil {
		jsonResponse(w, false, err.Error())
		return
	}

	jsonResponse(w, true, "")
}

func handleCommand(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CommandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResponse(w, false, err.Error())
		return
	}

	if err := input.SimulateCommand(req.Command); err != nil {
		jsonResponse(w, false, err.Error())
		return
	}

	jsonResponse(w, true, "")
}

func jsonResponse(w http.ResponseWriter, success bool, errMsg string) {
	w.Header().Set("Content-Type", "application/json")
	resp := Response{Success: success, Error: errMsg}
	json.NewEncoder(w).Encode(resp)
}
