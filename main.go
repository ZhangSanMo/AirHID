package main

import (
	"bufio"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/atotto/clipboard"
	"github.com/mdp/qrterminal/v3"
	"github.com/micmonay/keybd_event"
	"golang.org/x/sys/windows"
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

type MouseRequest struct {
	Action string  `json:"action"` // move, click, right_click, scroll
	X      float64 `json:"x"`      // dx for move
	Y      float64 `json:"y"`      // dy for move or scroll
}

var (
	user32         = syscall.NewLazyDLL("user32.dll")
	procMouseEvent = user32.NewProc("mouse_event")
)

const (
	MOUSEEVENTF_MOVE      = 0x0001
	MOUSEEVENTF_LEFTDOWN  = 0x0002
	MOUSEEVENTF_LEFTUP    = 0x0004
	MOUSEEVENTF_RIGHTDOWN = 0x0008
	MOUSEEVENTF_RIGHTUP   = 0x0010
	MOUSEEVENTF_WHEEL     = 0x0800
)

func main() {
	if !isAdmin() {
		fmt.Println("Warning: Running without Administrator privileges. Input simulation into elevated windows may fail.")
	}

	// Logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Get all available IPs and let user choose
	localIP := selectIP()
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
	http.HandleFunc("/mouse", handleMouse)
	http.HandleFunc("/api/info", handleInfo)

	log.Fatal(http.ListenAndServe("0.0.0.0:5000", nil))
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

	switch req.Action {
	case "move":
		mouseEvent(MOUSEEVENTF_MOVE, uintptr(int32(req.X)), uintptr(int32(req.Y)), 0, 0)
	case "click":
		mouseEvent(MOUSEEVENTF_LEFTDOWN, 0, 0, 0, 0)
		time.Sleep(10 * time.Millisecond)
		mouseEvent(MOUSEEVENTF_LEFTUP, 0, 0, 0, 0)
	case "right_click":
		mouseEvent(MOUSEEVENTF_RIGHTDOWN, 0, 0, 0, 0)
		time.Sleep(10 * time.Millisecond)
		mouseEvent(MOUSEEVENTF_RIGHTUP, 0, 0, 0, 0)
	case "scroll":
		// dwData for wheel is amount of movement in multiples of WHEEL_DELTA (120)
		mouseEvent(MOUSEEVENTF_WHEEL, 0, 0, uintptr(int32(req.Y)), 0)
	}

	jsonResponse(w, true, "")
}

func mouseEvent(dwFlags, dx, dy, dwData, dwExtraInfo uintptr) {
	procMouseEvent.Call(dwFlags, dx, dy, dwData, dwExtraInfo)
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
	case "left":
		kb.SetKeys(keybd_event.VK_LEFT)
	case "right":
		kb.SetKeys(keybd_event.VK_RIGHT)
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

// IPInfo stores IP address and its interface name
type IPInfo struct {
	IP        string
	Interface string
}

// getAllIPs returns all available IPv4 addresses with their interface names
func getAllIPs() []IPInfo {
	var ips []IPInfo
	
	interfaces, err := net.Interfaces()
	if err != nil {
		return ips
	}
	
	for _, iface := range interfaces {
		// Skip loopback and down interfaces
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}
		
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			
			// Only include IPv4 addresses and skip loopback
			if ip == nil || ip.IsLoopback() || ip.To4() == nil {
				continue
			}
			
			ips = append(ips, IPInfo{
				IP:        ip.String(),
				Interface: iface.Name,
			})
		}
	}
	
	return ips
}

// selectIP displays available IPs and lets user choose one
func selectIP() string {
	ips := getAllIPs()
	
	if len(ips) == 0 {
		fmt.Println("No network interface found, using localhost")
		return "localhost"
	}
	
	if len(ips) == 1 {
		fmt.Printf("Using IP: %s (%s)\n", ips[0].IP, ips[0].Interface)
		return ips[0].IP
	}
	
	// Multiple IPs available, let user choose
	fmt.Println("\nMultiple network interfaces detected:")
	fmt.Println(strings.Repeat("-", 50))
	for i, info := range ips {
		fmt.Printf("  [%d] %s (%s)\n", i+1, info.IP, info.Interface)
	}
	fmt.Println(strings.Repeat("-", 50))
	fmt.Printf("Please select IP [1-%d] (default 1): ", len(ips))
	
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	
	// Default to first IP if empty input
	if input == "" {
		return ips[0].IP
	}
	
	// Parse user selection
	choice, err := strconv.Atoi(input)
	if err != nil || choice < 1 || choice > len(ips) {
		fmt.Printf("Invalid selection, using default: %s\n", ips[0].IP)
		return ips[0].IP
	}
	
	return ips[choice-1].IP
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
