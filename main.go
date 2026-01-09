package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/atotto/clipboard"
	"github.com/mdp/qrterminal/v3"
	"github.com/micmonay/keybd_event"
	"golang.org/x/sys/windows"
)

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

func main() {
	if !isAdmin() {
		fmt.Println("Warning: Running without Administrator privileges. Input simulation into elevated windows may fail.")
	}

	// Logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Get Local IP
	localIP := getLocalIP()
	url := fmt.Sprintf("http://%s:5000", localIP)
	
	fmt.Printf("\n%s\n", strings.Repeat("=", 40))
	fmt.Printf("AirHID Running (Go Version)\n")
	fmt.Printf("Please visit: %s\n", url)
	fmt.Printf("%s\n\n", strings.Repeat("=", 40))

	// Generate QR Code
	qrterminal.GenerateHalfBlock(url, qrterminal.L, os.Stdout)
	fmt.Println()

	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/type", handleType)
	http.HandleFunc("/key", handleKey)
	http.HandleFunc("/api/info", handleInfo)

	log.Fatal(http.ListenAndServe("0.0.0.0:5000", nil))
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	tmplPath := filepath.Join("templates", "index.html")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		http.Error(w, "Template not found: "+err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

func handleInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "online", "version": "airhid-1.0"})
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

	if req.Mode == "type" {
		log.Printf("Injecting text via Clipboard: %.50s...", req.Text)
		if err := clipboard.WriteAll(req.Text); err != nil {
			jsonResponse(w, false, "Clipboard error: "+err.Error())
			return
		}
		
		// Simulate Ctrl+V
		time.Sleep(100 * time.Millisecond)
		if err := pressCtrlV(); err != nil {
			jsonResponse(w, false, "Key simulation error: "+err.Error())
			return
		}
		log.Println("Injected successfully")

	} else if req.Mode == "clipboard" {
		if err := clipboard.WriteAll(req.Text); err != nil {
			jsonResponse(w, false, "Clipboard error: "+err.Error())
			return
		}
	}

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

	log.Printf("Simulated Key: %s", req.Key)
	
	kb, err := keybd_event.NewKeyBonding()
	if err != nil {
		jsonResponse(w, false, "Key bonding error: "+err.Error())
		return
	}

	// For Linux/Windows specific set up if needed, but defaults are usually fine for standard keys
	// keybd_event defaults to windows on windows

	switch req.Key {
	case "ctrl_enter":
		kb.HasCTRL(true)
		kb.SetKeys(keybd_event.VK_ENTER)
	case "enter":
		kb.SetKeys(keybd_event.VK_ENTER)
	case "tab":
		kb.SetKeys(keybd_event.VK_TAB)
	case "backspace":
		kb.SetKeys(keybd_event.VK_BACK)
	case "esc":
		kb.SetKeys(keybd_event.VK_ESC)
	case "space":
		kb.SetKeys(keybd_event.VK_SPACE)
	case "up":
		kb.SetKeys(keybd_event.VK_UP)
	case "down":
		kb.SetKeys(keybd_event.VK_DOWN)
	default:
		// Try to map single characters if possible, but the API seems to send specific commands
		// If it sends 'a', 'b', etc., we would need a map. 
		// The python code had a fallback to `pyautogui.press(key)`.
		// For now we handle the buttons in the UI.
		jsonResponse(w, false, "Unknown key: "+req.Key)
		return
	}

	if err := kb.Launching(); err != nil {
		jsonResponse(w, false, "Key launch error: "+err.Error())
		return
	}

	jsonResponse(w, true, "")
}

func pressCtrlV() error {
	kb, err := keybd_event.NewKeyBonding()
	if err != nil {
		return err
	}
	kb.HasCTRL(true)
	kb.SetKeys(keybd_event.VK_V)
	return kb.Launching()
}

func jsonResponse(w http.ResponseWriter, success bool, errMsg string) {
	w.Header().Set("Content-Type", "application/json")
	resp := Response{Success: success, Error: errMsg}
	json.NewEncoder(w).Encode(resp)
}

func getLocalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "localhost"
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

func isAdmin() bool {
	// A simple way to check is to try to open the physical drive
	// This requires admin privileges on Windows
	_, err := os.Open("\\\\.\\PHYSICALDRIVE0")
	return err == nil
}

func runMeElevated() {
	verb := "runas"
	exe, _ := os.Executable()
	cwd, _ := os.Getwd()
	args := strings.Join(os.Args[1:], " ")

	verbPtr, _ := syscall.UTF16PtrFromString(verb)
	exePtr, _ := syscall.UTF16PtrFromString(exe)
	cwdPtr, _ := syscall.UTF16PtrFromString(cwd)
	argsPtr, _ := syscall.UTF16PtrFromString(args)

	var showCmd int32 = 1 // SW_NORMAL

	windows.ShellExecute(0, verbPtr, exePtr, argsPtr, cwdPtr, showCmd)
	os.Exit(0)
}
