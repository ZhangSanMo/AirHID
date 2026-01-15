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
// 支持的语法:
//   - 基础命令: ctrl+c, alt+tab, f5
//   - 重复执行: repeat:5:backspace, 3x enter, ctrl+z*10
//   - 中文口语: "执行3次ctrl加z", "ctrl加z五次", "按三下回车"
func SimulateCommand(cmd string) error {
	kbMutex.Lock()
	defer kbMutex.Unlock()

	cmd = strings.ToLower(strings.TrimSpace(cmd))
	if cmd == "" {
		return fmt.Errorf("命令不能为空")
	}

	// 仅替换中文"加"为"+"，其他口语词汇由状态机自动忽略
	cmd = strings.ReplaceAll(cmd, "加", "+")

	// 检查是否是重复执行命令
	repeatCount, actualCmd, isRepeat := parseRepeatCommand(cmd)
	if isRepeat {
		log.Printf("重复执行命令: %s, 次数: %d", actualCmd, repeatCount)
		for i := 0; i < repeatCount; i++ {
			if err := executeSingleCommand(actualCmd); err != nil {
				return fmt.Errorf("第 %d 次执行失败: %w", i+1, err)
			}
			if i < repeatCount-1 {
				time.Sleep(50 * time.Millisecond) // 每次执行之间的间隔
			}
		}
		return nil
	}

	return executeSingleCommand(cmd)
}

// executeSingleCommand 执行单个命令
func executeSingleCommand(cmd string) error {
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