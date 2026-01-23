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
	MouseActionDown       = "down"
	MouseActionUp         = "up"
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
	mainKeys, modifiers, err := ParseCommand(cmd)
	if err != nil {
		return err
	}

	kb, err := keybd_event.NewKeyBonding()
	if err != nil {
		return err
	}

	for _, m := range modifiers {
		switch m {
		case "Ctrl":
			kb.HasCTRL(true)
		case "Shift":
			kb.HasSHIFT(true)
		case "Alt":
			kb.HasALT(true)
		case "Win":
			kb.HasSuper(true)
		}
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
