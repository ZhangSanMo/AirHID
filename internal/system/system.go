package system

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"

	"golang.org/x/sys/windows/registry"
)

const (
	autoStartKeyPath = `Software\Microsoft\Windows\CurrentVersion\Run`
	appName          = "AirHID"
	taskName         = "AirHID_AutoStart"
)

// IsAdmin checks if the current process has administrator privileges
func IsAdmin() bool {
	_, err := os.Open(`\\.\PHYSICALDRIVE0`)
	return err == nil
}

// OpenBrowser opens a URL in the default browser
func OpenBrowser(url string) {
	err := exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	if err != nil {
		log.Println("Failed to open browser:", err)
	}
}

// AutoStartEnabled checks if the scheduled task exists
func AutoStartEnabled() bool {
	// Check Task Scheduler
	cmd := exec.Command("schtasks", "/Query", "/TN", taskName)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if err := cmd.Run(); err == nil {
		return true
	}
	return false
}

// EnableAutoStart creates a Scheduled Task with highest privileges
func EnableAutoStart() error {
	exePath, err := os.Executable()
	if err != nil {
		return err
	}

	// Cleanup legacy registry key if it exists
	cleanRegistryAutoStart()

	// Quote path to handle spaces
	cmdPath := fmt.Sprintf("\"%s\"", exePath)

	// Create Task
	// /SC ONLOGON : Start on login
	// /RL HIGHEST : Run as Administrator
	// /F : Force overwrite
	cmd := exec.Command("schtasks", "/Create", "/TN", taskName, "/TR", cmdPath, "/SC", "ONLOGON", "/RL", "HIGHEST", "/F")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd.Run()
}

// DisableAutoStart removes the scheduled task
func DisableAutoStart() error {
	cleanRegistryAutoStart()
	cmd := exec.Command("schtasks", "/Delete", "/TN", taskName, "/F")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd.Run()
}

func cleanRegistryAutoStart() {
	k, err := registry.OpenKey(registry.CURRENT_USER, autoStartKeyPath, registry.SET_VALUE)
	if err != nil {
		return
	}
	defer k.Close()
	k.DeleteValue(appName)
}
