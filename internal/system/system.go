package system

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"golang.org/x/sys/windows"
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

// runElevated runs a command with Administrator privileges using the "runas" verb
func runElevated(command string, args ...string) error {
	verb, _ := windows.UTF16PtrFromString("runas")
	exe, _ := windows.UTF16PtrFromString(command)
	params, _ := windows.UTF16PtrFromString(strings.Join(args, " "))
	cwd, _ := windows.UTF16PtrFromString("")

	return windows.ShellExecute(0, verb, exe, params, cwd, windows.SW_NORMAL)
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

	if !IsAdmin() {
		return runElevated("schtasks", "/Create", "/TN", taskName, "/TR", cmdPath, "/SC", "ONLOGON", "/RL", "HIGHEST", "/IT", "/DELAY", "0000:30", "/F")
	}

	// Create Task
	// /SC ONLOGON : Start on login
	// /RL HIGHEST : Run as Administrator
	// /IT : Run only if user is logged on (Interactive) - Fixes visibility issues
	// /DELAY 0000:30 : Wait 30 seconds for system/network initialization
	// /F : Force overwrite
	cmd := exec.Command("schtasks", "/Create", "/TN", taskName, "/TR", cmdPath, "/SC", "ONLOGON", "/RL", "HIGHEST", "/IT", "/DELAY", "0000:30", "/F")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd.Run()
}

// DisableAutoStart removes the scheduled task
func DisableAutoStart() error {
	cleanRegistryAutoStart()

	if !IsAdmin() {
		return runElevated("schtasks", "/Delete", "/TN", taskName, "/F")
	}

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
