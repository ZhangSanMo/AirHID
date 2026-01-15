package system

import (
	"log"
	"os"
	"os/exec"

	"golang.org/x/sys/windows/registry"
)

const (
	autoStartKeyPath = `Software\Microsoft\Windows\CurrentVersion\Run`
	appName          = "AirHID"
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

// AutoStartEnabled checks if the registry key exists and points to current exe
func AutoStartEnabled() bool {
	k, err := registry.OpenKey(registry.CURRENT_USER, autoStartKeyPath, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer k.Close()

	val, _, err := k.GetStringValue(appName)
	if err != nil {
		return false
	}

	exePath, _ := os.Executable()
	return val == exePath
}

// EnableAutoStart adds the registry key
func EnableAutoStart() error {
	exePath, err := os.Executable()
	if err != nil {
		return err
	}

	k, err := registry.OpenKey(registry.CURRENT_USER, autoStartKeyPath, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()

	return k.SetStringValue(appName, exePath)
}

// DisableAutoStart removes the registry key
func DisableAutoStart() error {
	k, err := registry.OpenKey(registry.CURRENT_USER, autoStartKeyPath, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()

	return k.DeleteValue(appName)
}
