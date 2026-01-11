//go:build !windows

package input

import "log"

func SimulateMouse(action string, x, y float64) {
	log.Println("Mouse simulation is only supported on Windows.")
}
