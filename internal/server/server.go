package server

import (
	"crypto/rand"
	"embed"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/skip2/go-qrcode"

	"airhid/internal/input"
)

//go:embed templates/index.html
//go:embed templates/connect.html
var templatesFS embed.FS

type Response struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type ConnectData struct {
	URL    string
	QRCode string
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

var (
	serverToken string
	serverHost  string
	serverPort  string
	displayAddr string // IP to show in QR code
)

func SetupRoutes(host, port, token, displayIP string) {
	serverToken = token
	serverHost = host
	serverPort = port
	displayAddr = displayIP

	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/connect", handleConnect)

	// Protected routes
	http.HandleFunc("/type", authMiddleware(handleType))
	http.HandleFunc("/command", authMiddleware(handleCommand))
	http.HandleFunc("/key", authMiddleware(handleKey))
	http.HandleFunc("/mouse", authMiddleware(handleMouse))
	http.HandleFunc("/api/info", authMiddleware(handleInfo))
}

func Start() error {
	addr := fmt.Sprintf("%s:%s", serverHost, serverPort)
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

func handleConnect(w http.ResponseWriter, r *http.Request) {
	url := fmt.Sprintf("http://%s:%s/?token=%s", displayAddr, serverPort, serverToken)

	png, err := qrcode.Encode(url, qrcode.Medium, 256)
	if err != nil {
		http.Error(w, "Failed to generate QR code", http.StatusInternalServerError)
		return
	}

	data := ConnectData{
		URL:    url,
		QRCode: base64.StdEncoding.EncodeToString(png),
	}

	tmpl, err := template.ParseFS(templatesFS, "templates/connect.html")
	if err != nil {
		http.Error(w, "Template not found: "+err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, data)
}

func jsonResponse(w http.ResponseWriter, success bool, errMsg string) {
	w.Header().Set("Content-Type", "application/json")
	resp := Response{Success: success, Error: errMsg}
	json.NewEncoder(w).Encode(resp)
}
