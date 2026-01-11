package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"syscall"

	"github.com/mdp/qrterminal/v3"
	"golang.org/x/sys/windows"

	"airhid/internal/config"
	"airhid/internal/network"
	"airhid/internal/server"
)

func main() {
	if !isAdmin() {
		fmt.Println("Warning: Running without Administrator privileges. Input simulation into elevated windows may fail.")
	}

	// Logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// 1. Load Config
	cfg, err := config.LoadOrInit()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// 2. Determine Display IP
	displayIP := cfg.Host
	if cfg.Host == "0.0.0.0" {
		displayIP = network.GetDefaultIP()
	}

	// 3. Construct URL
	url := fmt.Sprintf("http://%s:%s/?token=%s", displayIP, cfg.Port, cfg.Token)

	// 4. UI Output
	fmt.Printf("\n%s\n", strings.Repeat("=", 40))
	fmt.Printf("AirHID Running (Secure Mode)\n")
	fmt.Printf("Listening on: %s:%s\n", cfg.Host, cfg.Port)
	fmt.Printf("Connect URL:  %s\n", url)
	
	// If listening on all interfaces (0.0.0.0), show other options
	if cfg.Host == "0.0.0.0" {
		allIPs := network.GetAllIPs()
		if len(allIPs) > 1 {
			fmt.Println("\nAlso available on:")
			for _, info := range allIPs {
				if info.IP != displayIP {
					fmt.Printf("  - http://%s:%s/?token=%s\n", info.IP, cfg.Port, cfg.Token)
				}
			}
		}
	}
	fmt.Printf("%s\n\n", strings.Repeat("=", 40))

	// Generate QR Code
	qrterminal.GenerateHalfBlock(url, qrterminal.L, os.Stdout)
	fmt.Println()

	// 5. Start Server
	if err := server.Start(cfg.Host, cfg.Port, cfg.Token); err != nil {
		log.Fatal(err)
	}
}

func isAdmin() bool {
	_, err := os.Open(`\\.\PHYSICALDRIVE0`)
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