package main

import (
	"fmt"
	"log"

	"airhid/internal/system"
	"airhid/internal/tray"
)

func main() {
	// Setup Logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Admin Check (Warning only)
	if !system.IsAdmin() {
		fmt.Println("Warning: Running as standard user.")
		fmt.Println("Input simulation may not work on elevated windows (e.g. Task Manager).")
		fmt.Println("Run as Administrator if you need full control.")
	}

	// Start Application
	tray.Run()
}