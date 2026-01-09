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
	"sync"
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

type CommandRequest struct {
	Command string `json:"command"`
}

type MouseRequest struct {
	Action string  `json:"action"` // move, click, right_click, scroll
	X      float64 `json:"x"`      // dx for move
	Y      float64 `json:"y"`      // dy for move or scroll
}

var (
	user32         = syscall.NewLazyDLL("user32.dll")
	procMouseEvent = user32.NewProc("mouse_event")
	kbMutex        sync.Mutex
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
	http.HandleFunc("/command", handleCommand)
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
		kbMutex.Lock()
		defer kbMutex.Unlock()
		
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

	kbMutex.Lock()
	defer kbMutex.Unlock()

	log.Printf("Simulated Key: %s", req.Key)
	
	kb, err := keybd_event.NewKeyBonding()
	if err != nil {
		jsonResponse(w, false, "Key bonding error: "+err.Error())
		return
	}

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
		jsonResponse(w, false, "Unknown key: "+req.Key)
		return
	}

	if err := kb.Launching(); err != nil {
		jsonResponse(w, false, "Key launch error: "+err.Error())
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

	kbMutex.Lock()
	defer kbMutex.Unlock()

	log.Printf("收到原始命令: %q", req.Command)
	err := executeCommand(req.Command)
	if err != nil {
		log.Printf("命令执行失败: %v", err)
		jsonResponse(w, false, err.Error())
		return
	}

	jsonResponse(w, true, "")
}

func executeCommand(cmd string) error {
	cmd = strings.ToLower(strings.TrimSpace(cmd))
	if cmd == "" {
		return fmt.Errorf("命令不能为空")
	}

	// 1. 准备按键识别基础
	modifierNames := []string{"control", "ctrl", "shift", "alt", "windows", "win", "command", "cmd", "meta", "super"}

	kb, err := keybd_event.NewKeyBonding()
	if err != nil {
		return err
	}

	var mainKeys []int
	var modifiers []string
	inSegment := false // 标记是否正处于一个有效按键段中

	runes := []rune(cmd)
	for i := 0; i < len(runes); {
		matchLen := 0
		foundKey := ""
		isMod := false
		currentSuffix := string(runes[i:])

		// A. 尝试匹配长名修饰键
		for _, m := range modifierNames {
			if strings.HasPrefix(currentSuffix, m) {
				foundKey = m
				isMod = true
				matchLen = len([]rune(m))
				break
			}
		}

		// B. 尝试匹配 keyMap 中的长按键名 (优先匹配最长的)
		if foundKey == "" {
			for k := range keyMap {
				if strings.HasPrefix(currentSuffix, k) {
					if len([]rune(k)) > matchLen {
						foundKey = k
						matchLen = len([]rune(k))
					}
				}
			}
			// 特殊处理空格描述
			if strings.HasPrefix(currentSuffix, "space") {
				if 5 > matchLen {
					foundKey = "space"
					matchLen = 5
				}
			}
			if strings.HasPrefix(currentSuffix, "空格") {
				if 2 > matchLen {
					foundKey = "空格"
					matchLen = 2
				}
			}
		}

		// C. 尝试匹配 charMap (单字符，排除字面空格)
		if foundKey == "" {
			r := runes[i]
			if r < 128 && r != ' ' { // ASCII 字符且不是普通空格
				if _, ok := charMap[byte(r)]; ok {
					foundKey = string(r)
					matchLen = 1
				}
			}
		}

		// 2. 状态机逻辑
		if foundKey != "" {
			// 如果当前不在有效段内，记录该段的第一个有效按键
			if !inSegment {
				if isMod {
					switch foundKey {
					case "ctrl", "control":
						kb.HasCTRL(true)
						modifiers = append(modifiers, "Ctrl")
					case "shift":
						kb.HasSHIFT(true)
						modifiers = append(modifiers, "Shift")
					case "alt":
						kb.HasALT(true)
						modifiers = append(modifiers, "Alt")
					case "win", "windows", "command", "cmd", "meta", "super":
						kb.HasSuper(true)
						modifiers = append(modifiers, "Win")
					}
				} else {
					if foundKey == "space" || foundKey == "空格" {
						mainKeys = append(mainKeys, keybd_event.VK_SPACE)
					} else if vk, ok := keyMap[foundKey]; ok {
						mainKeys = append(mainKeys, vk)
					} else {
						mainKeys = append(mainKeys, charMap[foundKey[0]])
					}
				}
				inSegment = true // 进入有效段，后续在遇到分隔符前的按键都将忽略
			}
			i += matchLen
		} else {
			// 当前字符无法识别，视为分隔符，结束当前有效段
			inSegment = false
			i++
		}
	}

	if len(mainKeys) == 0 && len(modifiers) == 0 {
		return fmt.Errorf("未能识别出有效按键: %s", cmd)
	}

	log.Printf("解析结果 -> 修饰键: %v, 按键序列: %v", modifiers, mainKeys)

	// 3. 执行模拟
	if len(modifiers) > 0 || len(mainKeys) <= 1 {
		kb.SetKeys(mainKeys...)
		return kb.Launching()
	} else {
		for _, k := range mainKeys {
			singleKb, _ := keybd_event.NewKeyBonding()
			singleKb.SetKeys(k)
			if err := singleKb.Launching(); err != nil {
				return err
			}
			time.Sleep(10 * time.Millisecond)
		}
	}

	return nil
}

var keyMap = map[string]int{
	// 基础控制
	"enter":     keybd_event.VK_ENTER,
	"回车":       keybd_event.VK_ENTER,
	"确认":       keybd_event.VK_ENTER,
	"esc":       keybd_event.VK_ESC,
	"escape":    keybd_event.VK_ESC,
	"退出":       keybd_event.VK_ESC,
	"tab":       keybd_event.VK_TAB,
	"制表":       keybd_event.VK_TAB,
	"space":     keybd_event.VK_SPACE,
	"空格":       keybd_event.VK_SPACE,
	"backspace": keybd_event.VK_BACK,
	"退格":       keybd_event.VK_BACK,
	"del":       keybd_event.VK_DELETE,
	"delete":    keybd_event.VK_DELETE,
	"删除":       keybd_event.VK_DELETE,
	"ins":       keybd_event.VK_INSERT,
	"insert":    keybd_event.VK_INSERT,
	"插入":       keybd_event.VK_INSERT,
	"prtsc":     0x2C,
	"printscreen": 0x2C,
	"截屏":       0x2C,

	// 方向键
	"up":    keybd_event.VK_UP,
	"down":  keybd_event.VK_DOWN,
	"left":  keybd_event.VK_LEFT,
	"right": keybd_event.VK_RIGHT,
	"上":     keybd_event.VK_UP,
	"下":     keybd_event.VK_DOWN,
	"左":     keybd_event.VK_LEFT,
	"右":     keybd_event.VK_RIGHT,

	// 功能键
	"f1": keybd_event.VK_F1, "f2": keybd_event.VK_F2, "f3": keybd_event.VK_F3, "f4": keybd_event.VK_F4,
	"f5": keybd_event.VK_F5, "f6": keybd_event.VK_F6, "f7": keybd_event.VK_F7, "f8": keybd_event.VK_F8,
	"f9": keybd_event.VK_F9, "f10": keybd_event.VK_F10, "f11": keybd_event.VK_F11, "f12": keybd_event.VK_F12,

	// 页面控制
	"home":   keybd_event.VK_HOME,
	"end":    keybd_event.VK_END,
	"pgup":   keybd_event.VK_PAGEUP,
	"pgdn":   keybd_event.VK_PAGEDOWN,
	"pageup": keybd_event.VK_PAGEUP,
	"pagedown": keybd_event.VK_PAGEDOWN,
	"向上翻页": keybd_event.VK_PAGEUP,
	"向下翻页": keybd_event.VK_PAGEDOWN,

	// 符号映射 (Windows VK Codes)
	"+": 0xBB, "加号": 0xBB,
	"-": 0xBD, "减号": 0xBD,
	"=": 0xBB, "等于": 0xBB,
	",": 0xBC, "逗号": 0xBC,
	".": 0xBE, "句号": 0xBE,
	"/": 0xBF, "斜杠": 0xBF,
	";": 0xBA, "分号": 0xBA,
	"'": 0xDE, "引号": 0xDE,
	"[": 0xDB, "左括号": 0xDB,
	"]": 0xDD, "右括号": 0xDD,
	"\\": 0xdc, "反斜杠": 0xdc,
	"`": 0xc0, "波浪号": 0xc0,
}

var charMap = map[byte]int{
	'a': keybd_event.VK_A, 'b': keybd_event.VK_B, 'c': keybd_event.VK_C, 'd': keybd_event.VK_D,
	'e': keybd_event.VK_E, 'f': keybd_event.VK_F, 'g': keybd_event.VK_G, 'h': keybd_event.VK_H,
	'i': keybd_event.VK_I, 'j': keybd_event.VK_J, 'k': keybd_event.VK_K, 'l': keybd_event.VK_L,
	'm': keybd_event.VK_M, 'n': keybd_event.VK_N, 'o': keybd_event.VK_O, 'p': keybd_event.VK_P,
	'q': keybd_event.VK_Q, 'r': keybd_event.VK_R, 's': keybd_event.VK_S, 't': keybd_event.VK_T,
	'u': keybd_event.VK_U, 'v': keybd_event.VK_V, 'w': keybd_event.VK_W, 'x': keybd_event.VK_X,
	'y': keybd_event.VK_Y, 'z': keybd_event.VK_Z,
	'0': keybd_event.VK_0, '1': keybd_event.VK_1, '2': keybd_event.VK_2, '3': keybd_event.VK_3,
	'4': keybd_event.VK_4, '5': keybd_event.VK_5, '6': keybd_event.VK_6, '7': keybd_event.VK_7,
	'8': keybd_event.VK_8, '9': keybd_event.VK_9,
	' ': keybd_event.VK_SPACE,
	'+': 0xBB, '-': 0xBD, '=': 0xBB, ',': 0xBC, '.': 0xBE, '/': 0xBF,
	';': 0xBA, '\'': 0xDE, '[': 0xDB, ']': 0xDD, '\\': 0xDC, '`': 0xC0,
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
