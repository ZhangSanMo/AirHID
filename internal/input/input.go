package input

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/atotto/clipboard"
	"github.com/micmonay/keybd_event"
)

var (
	kbMutex sync.Mutex
)

// MouseAction 定义鼠标动作
const (
	MouseActionMove       = "move"
	MouseActionClick      = "click"
	MouseActionRightClick = "right_click"
	MouseActionScroll     = "scroll"
)

// SimulateType 模拟文本输入
func SimulateType(text, mode string) error {
	if mode == "type" {
		kbMutex.Lock()
		defer kbMutex.Unlock()

		log.Printf("Injecting text via Clipboard: %.50s...", text)
		if err := clipboard.WriteAll(text); err != nil {
			return fmt.Errorf("clipboard error: %w", err)
		}

		// Simulate Ctrl+V
		time.Sleep(100 * time.Millisecond)
		if err := pressCtrlV(); err != nil {
			return fmt.Errorf("key simulation error: %w", err)
		}
		log.Println("Injected successfully")
		return nil

	} else if mode == "clipboard" {
		if err := clipboard.WriteAll(text); err != nil {
			return fmt.Errorf("clipboard error: %w", err)
		}
		return nil
	}
	return nil
}

// SimulateKey 模拟单个按键
func SimulateKey(key string) error {
	kbMutex.Lock()
	defer kbMutex.Unlock()

	log.Printf("Simulated Key: %s", key)

	kb, err := keybd_event.NewKeyBonding()
	if err != nil {
		return fmt.Errorf("key bonding error: %w", err)
	}

	switch key {
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
		return fmt.Errorf("unknown key: %s", key)
	}

	if err := kb.Launching(); err != nil {
		return fmt.Errorf("key launch error: %w", err)
	}
	return nil
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

// SimulateCommand 解析并执行复杂命令
func SimulateCommand(cmd string) error {
	kbMutex.Lock()
	defer kbMutex.Unlock()

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
	inSegment := false

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

		// B. 尝试匹配 keyMap
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

		// C. 尝试匹配 charMap
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
				inSegment = true
			}
			i += matchLen
		} else {
			inSegment = false
			i++
		}
	}

	if len(mainKeys) == 0 && len(modifiers) == 0 {
		return fmt.Errorf("未能识别出有效按键: %s", cmd)
	}

	log.Printf("解析结果 -> 修饰键: %v, 按键序列: %v", modifiers, mainKeys)

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

// 映射表移动到这里
var keyMap = map[string]int{
	"enter": keybd_event.VK_ENTER, "回车": keybd_event.VK_ENTER, "确认": keybd_event.VK_ENTER,
	"esc": keybd_event.VK_ESC, "escape": keybd_event.VK_ESC, "退出": keybd_event.VK_ESC,
	"tab": keybd_event.VK_TAB, "制表": keybd_event.VK_TAB,
	"space": keybd_event.VK_SPACE, "空格": keybd_event.VK_SPACE,
	"backspace": keybd_event.VK_BACK, "退格": keybd_event.VK_BACK,
	"del": keybd_event.VK_DELETE, "delete": keybd_event.VK_DELETE, "删除": keybd_event.VK_DELETE,
	"ins": keybd_event.VK_INSERT, "insert": keybd_event.VK_INSERT, "插入": keybd_event.VK_INSERT,
	"prtsc": 0x2C, "printscreen": 0x2C, "截屏": 0x2C,
	"up": keybd_event.VK_UP, "down": keybd_event.VK_DOWN, "left": keybd_event.VK_LEFT, "right": keybd_event.VK_RIGHT,
	"上": keybd_event.VK_UP, "下": keybd_event.VK_DOWN, "左": keybd_event.VK_LEFT, "右": keybd_event.VK_RIGHT,
	"home": keybd_event.VK_HOME, "end": keybd_event.VK_END,
	"pgup": keybd_event.VK_PAGEUP, "pgdn": keybd_event.VK_PAGEDOWN,
	"pageup": keybd_event.VK_PAGEUP, "pagedown": keybd_event.VK_PAGEDOWN,
	"向上翻页": keybd_event.VK_PAGEUP, "向下翻页": keybd_event.VK_PAGEDOWN,
	"f1": keybd_event.VK_F1, "f2": keybd_event.VK_F2, "f3": keybd_event.VK_F3, "f4": keybd_event.VK_F4,
	"f5": keybd_event.VK_F5, "f6": keybd_event.VK_F6, "f7": keybd_event.VK_F7, "f8": keybd_event.VK_F8,
	"f9": keybd_event.VK_F9, "f10": keybd_event.VK_F10, "f11": keybd_event.VK_F11, "f12": keybd_event.VK_F12,
	"+": 0xBB, "加号": 0xBB, "-": 0xBD, "减号": 0xBD, "=": 0xBB, "等于": 0xBB,
	",": 0xBC, "逗号": 0xBC, ".": 0xBE, "句号": 0xBE, "/": 0xBF, "斜杠": 0xBF,
	";": 0xBA, "分号": 0xBA, "'": 0xDE, "引号": 0xDE,
	"[": 0xDB, "左括号": 0xDB, "]": 0xDD, "右括号": 0xDD, "\\": 0xdc, "反斜杠": 0xdc, "`": 0xc0, "波浪号": 0xc0,
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
