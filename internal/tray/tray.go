package tray

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/getlantern/systray"

	"airhid/internal/config"
	"airhid/internal/icon"
	"airhid/internal/network"
	"airhid/internal/server"
	"airhid/internal/system"
)

// Run starts the system tray and blocks until exit
func Run() {
	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetIcon(icon.Data())
	systray.SetTitle("AirHID")
	systray.SetTooltip("AirHID Server")

	mStatus := systray.AddMenuItem("Server: Starting...", "Server Status")
	mInfo := systray.AddMenuItem("Show Connection Info", "Open the connection page with QR code")
	systray.AddSeparator()
	mAutoStart := systray.AddMenuItem("Enable Auto-Start", "Start AirHID when you log in")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Exit Application")

	// Start local server
	go func() {
		mStatus.SetTitle("Server: Running")
		err := runServer()
		if err != nil {
			log.Println("Local server start failed:", err)
			mStatus.SetTitle("Server: Stopped (Port Busy?)")
		}
	}()

	// Periodic status check for Auto-Start
	go func() {
		ticker := time.NewTicker(1 * time.Second) // Check immediately then loop
		defer ticker.Stop()
		for {
			enabled := system.AutoStartEnabled()
			if enabled {
				mAutoStart.SetTitle("Disable Auto-Start")
				mAutoStart.Check()
			} else {
				mAutoStart.SetTitle("Enable Auto-Start")
				mAutoStart.Uncheck()
			}
			
			// Wait for tick or click
			select {
			case <-ticker.C:
				continue
			case <-mAutoStart.ClickedCh:
				if enabled {
					if err := system.DisableAutoStart(); err != nil {
						log.Println("Error disabling auto-start:", err)
					}
				} else {
					if err := system.EnableAutoStart(); err != nil {
						log.Println("Error enabling auto-start:", err)
					}
				}
			case <-mInfo.ClickedCh:
				cfg, _ := config.LoadOrInit()
				if cfg != nil {
					ip := cfg.Host
					if ip == "0.0.0.0" {
						ip = network.GetDefaultIP()
					}
					url := fmt.Sprintf("http://%s:%s/connect", ip, cfg.Port)
					system.OpenBrowser(url)
				}
			case <-mQuit.ClickedCh:
				systray.Quit()
			}
		}
	}()
}

func onExit() {
	// Cleanup if needed
}

func runServer() error {
	cfg, err := config.LoadOrInit()
	if err != nil {
		return err
	}

	// For console output (if visible)
	displayIP := cfg.Host
	if cfg.Host == "0.0.0.0" {
		displayIP = network.GetDefaultIP()
	}
	// Ensure config dir is correct relative to exe
	exePath, _ := os.Executable()
	os.Chdir(filepath.Dir(exePath))

	url := fmt.Sprintf("http://%s:%s/?token=%s", displayIP, cfg.Port, cfg.Token)
	
	log.Printf("Starting AirHID on %s:%s", cfg.Host, cfg.Port)
	log.Printf("URL: %s", url)

	return server.Start(cfg.Host, cfg.Port, cfg.Token, displayIP)
}
