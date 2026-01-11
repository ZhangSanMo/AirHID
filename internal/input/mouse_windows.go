package input

import (
	"syscall"
	"time"
)

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

func SimulateMouse(action string, x, y float64) {
	switch action {
	case MouseActionMove:
		mouseEvent(MOUSEEVENTF_MOVE, uintptr(int32(x)), uintptr(int32(y)), 0, 0)
	case MouseActionClick:
		mouseEvent(MOUSEEVENTF_LEFTDOWN, 0, 0, 0, 0)
		time.Sleep(10 * time.Millisecond)
		mouseEvent(MOUSEEVENTF_LEFTUP, 0, 0, 0, 0)
	case MouseActionRightClick:
		mouseEvent(MOUSEEVENTF_RIGHTDOWN, 0, 0, 0, 0)
		time.Sleep(10 * time.Millisecond)
		mouseEvent(MOUSEEVENTF_RIGHTUP, 0, 0, 0, 0)
	case MouseActionScroll:
		mouseEvent(MOUSEEVENTF_WHEEL, 0, 0, uintptr(int32(y)), 0)
	}
}

func mouseEvent(dwFlags, dx, dy, dwData, dwExtraInfo uintptr) {
	procMouseEvent.Call(dwFlags, dx, dy, dwData, dwExtraInfo)
}
