package network

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

// IPInfo stores IP address and its interface name
type IPInfo struct {
	IP        string
	Interface string
}

// GetAllIPs returns all available IPv4 addresses with their interface names
func GetAllIPs() []IPInfo {
	var ips []IPInfo
	
	interfaces, err := net.Interfaces()
	if err != nil {
		return ips
	}
	
	for _, iface := range interfaces {
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}
		
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
		
			if ip == nil || ip.IsLoopback() || ip.To4() == nil {
				continue
			}
		
			ips = append(ips, IPInfo{
				IP:        ip.String(),
				Interface: iface.Name,
			})
		}
	}
	
	return ips
}

// SelectIP displays available IPs and lets user choose one
func SelectIP() string {
	ips := GetAllIPs()
	
	if len(ips) == 0 {
		fmt.Println("No network interface found, using localhost")
		return "localhost"
	}
	
	if len(ips) == 1 {
		fmt.Printf("Using IP: %s (%s)\n", ips[0].IP, ips[0].Interface)
		return ips[0].IP
	}
	
	fmt.Println("\nMultiple network interfaces detected:")
	fmt.Println(strings.Repeat("-", 50))
	for i, info := range ips {
		fmt.Printf("  [%d] %s (%s)\n", i+1, info.IP, info.Interface)
	}
	fmt.Println(strings.Repeat("-", 50))
	fmt.Printf("Please select IP [1-%d] (default 1): ", len(ips))
	
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	
	if input == "" {
		return ips[0].IP
	}
	
	choice, err := strconv.Atoi(input)
	if err != nil || choice < 1 || choice > len(ips) {
		fmt.Printf("Invalid selection, using default: %s\n", ips[0].IP)
		return ips[0].IP
	}
	
	return ips[choice-1].IP
}

// GetDefaultIP automatically picks the most likely local IP address
func GetDefaultIP() string {
	ips := GetAllIPs()
	if len(ips) == 0 {
		return "localhost"
	}
	
	// Priority: 192.168.x.x > 10.x.x.x > others
	for _, info := range ips {
		if strings.HasPrefix(info.IP, "192.168.") {
			return info.IP
		}
	}
	for _, info := range ips {
		if strings.HasPrefix(info.IP, "10.") {
			return info.IP
		}
	}
	
	return ips[0].IP
}
