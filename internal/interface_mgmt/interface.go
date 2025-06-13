package interface_mgmt

import (
	"fmt"
	"net"
	"strings"

	"github.com/RimakiTaema/netmgr/internal/common"
)

// Interface represents a network interface
type Interface struct {
	Name      string `json:"name"`
	IPAddress string `json:"ip_address"`
	Netmask   string `json:"netmask"`
	Gateway   string `json:"gateway"`
	MAC       string `json:"mac"`
	MTU       int    `json:"mtu"`
	State     string `json:"state"`
}

// HandleCommand processes interface management commands
func HandleCommand(command string, params []string) {
	switch command {
	case "show", "list":
		if len(params) > 0 {
			showInterface(params[0])
		} else {
			showAllInterfaces()
		}
	case "set", "config":
		if len(params) < 3 {
			common.LogError("Usage: interface set <interface> <property> <value> [<value2>]")
			return
		}
		setInterface(params[0], params[1], params[2:])
	case "help":
		showHelp()
	default:
		common.LogError("Unknown interface command: %s", command)
	}
}

func showHelp() {
	fmt.Println(`Interface Management Commands:

    show [interface]       Show all interfaces or details of specific interface
    set <interface> <property> <value> [<value2>]
                          Configure interface properties
                          
Properties:
    ip <address> [prefix]  Set IP address with optional prefix length
    up                     Bring interface up
    down                   Bring interface down
    mtu <size>             Set MTU size
    mac <address>          Set MAC address

Examples:
    netmgr interface show eth0
    netmgr interface set eth0 ip 192.168.1.100 24
    netmgr interface set eth0 up
    netmgr interface set eth0 mtu 1500`)
}

func showAllInterfaces() {
	common.LogInfo("All network interfaces:")
	fmt.Println()

	// Header
	fmt.Printf("%-15s %-10s %-15s %-20s %-10s\n", "INTERFACE", "STATE", "IP ADDRESS", "MAC ADDRESS", "MTU")
	fmt.Printf("%-15s %-10s %-15s %-20s %-10s\n", "---------", "-----", "----------", "-----------", "---")

	// Get all interfaces
	interfaces, err := net.Interfaces()
	if err != nil {
		common.LogError("Failed to get interfaces: %v", err)
		return
	}

	for _, iface := range interfaces {
		// Skip loopback
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		// Get state
		state := "DOWN"
		if iface.Flags&net.FlagUp != 0 {
			state = "UP"
		}

		// Get IP address
		addrs, err := iface.Addrs()
		ipAddr := "N/A"
		if err == nil && len(addrs) > 0 {
			for _, addr := range addrs {
				if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
					if ipnet.IP.To4() != nil {
						ipAddr = ipnet.IP.String()
						break
					}
				}
			}
		}

		// Format MAC address
		mac := iface.HardwareAddr.String()
		if mac == "" {
			mac = "N/A"
		}

		fmt.Printf("%-15s %-10s %-15s %-20s %-10d\n", iface.Name, state, ipAddr, mac, iface.MTU)
	}
}

func showInterface(interfaceName string) {
	common.LogInfo("Interface details for: %s", interfaceName)
	fmt.Println()

	// Find the interface
	iface, err := net.InterfaceByName(interfaceName)
	if err != nil {
		common.LogError("Interface %s not found: %v", interfaceName, err)
		return
	}

	// Basic info
	fmt.Printf("%s=== Interface Information ===%s\n", common.ColorCyan, common.ColorReset)
	fmt.Printf("Name: %s\n", iface.Name)
	fmt.Printf("Index: %d\n", iface.Index)
	fmt.Printf("MTU: %d\n", iface.MTU)
	fmt.Printf("Hardware Address: %s\n", iface.HardwareAddr)
	
	// State
	state := "DOWN"
	if iface.Flags&net.FlagUp != 0 {
		state = "UP"
	}
	fmt.Printf("State: %s\n", state)
	fmt.Printf("Flags: %s\n", iface.Flags.String())

	// IP addresses
	fmt.Printf("\nIP Addresses:\n")
	addrs, err := iface.Addrs()
	if err != nil {
		common.LogError("Failed to get addresses: %v", err)
	} else {
		for _, addr := range addrs {
			fmt.Printf("  %s\n", addr.String())
		}
	}

	// Platform-specific details
	if common.IsLinux() {
		showLinuxInterfaceDetails(interfaceName)
	} else if common.IsWindows() {
		showWindowsInterfaceDetails(interfaceName)
	} else if common.IsMacOS() {
		showMacOSInterfaceDetails(interfaceName)
	}
}

func showLinuxInterfaceDetails(interfaceName string) {
	// Statistics
	fmt.Printf("\n%s=== Statistics ===%s\n", common.ColorCyan, common.ColorReset)
	output, err := common.Execute("ip", "-s", "link", "show", interfaceName)
	if err == nil {
		fmt.Println(output)
	}

	// Routes
	fmt.Printf("\n%s=== Routes ===%s\n", common.ColorCyan, common.ColorReset)
	output, err = common.Execute("ip", "route", "show", "dev", interfaceName)
	if err == nil {
		fmt.Println(output)
	}

	// Active connections
	fmt.Printf("\n%s=== Active Connections ===%s\n", common.ColorCyan, common.ColorReset)
	output, err = common.Execute("ss", "-i")
	if err == nil {
		lines := strings.Split(output, "\n")
		found := false
		for _, line := range lines {
			if strings.Contains(line, interfaceName) {
				fmt.Println(line)
				found = true
			}
		}
		if !found {
			fmt.Println("No active connections")
		}
	}
}

func showWindowsInterfaceDetails(interfaceName string) {
	// Get interface details using netsh
	fmt.Printf("\n%s=== Interface Details ===%s\n", common.ColorCyan, common.ColorReset)
	output, err := common.Execute("netsh", "interface", "ip", "show", "addresses", interfaceName)
	if err == nil {
		fmt.Println(output)
	}

	// Get statistics
	fmt.Printf("\n%s=== Statistics ===%s\n", common.ColorCyan, common.ColorReset)
	output, err = common.Execute("netsh", "interface", "ip", "show", "interface", interfaceName)
	if err == nil {
		fmt.Println(output)
	}

	// Get routes
	fmt.Printf("\n%s=== Routes ===%s\n", common.ColorCyan, common.ColorReset)
	output, err = common.Execute("route", "print", "-4", interfaceName)
	if err == nil {
		fmt.Println(output)
	}
}

func showMacOSInterfaceDetails(interfaceName string) {
	// Get interface details
	fmt.Printf("\n%s=== Interface Details ===%s\n", common.ColorCyan, common.ColorReset)
	output, err := common.Execute("ifconfig", interfaceName)
	if err == nil {
		fmt.Println(output)
	}

	// Get routes
	fmt.Printf("\n%s=== Routes ===%s\n", common.ColorCyan, common.ColorReset)
	output, err = common.Execute("netstat", "-rn")
	if err == nil {
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			if strings.Contains(line, interfaceName) {
				fmt.Println(line)
			}
		}
	}
}

func setInterface(interfaceName, property string, values []string) {
	switch property {
	case "ip", "address":
		if len(values) < 1 {
			common.LogError("Usage: interface set <interface> ip <address> [prefix]")
			return
		}
		ipAddress := values[0]
		prefix := "24" // Default prefix
		if len(values) > 1 {
			prefix = values[1]
		}
		setInterfaceIP(interfaceName, ipAddress, prefix)
	case "up":
		setInterfaceState(interfaceName, true)
	case "down":
		setInterfaceState(interfaceName, false)
	case "mtu":
		if len(values) < 1 {
			common.LogError("Usage: interface set <interface> mtu <size>")
			return
		}
		setInterfaceMTU(interfaceName, values[0])
	case "mac":
		if len(values) < 1 {
			common.LogError("Usage: interface set <interface> mac <address>")
			return
		}
		setInterfaceMAC(interfaceName, values[0])
	default:
		common.LogError("Unknown property: %s", property)
	}
}

func setInterfaceIP(interfaceName, ipAddress, prefix string) {
	common.LogInfo("Setting IP address %s/%s on %s", ipAddress, prefix, interfaceName)

	if common.IsLinux() {
		_, err := common.Execute("ip", "addr", "add", ipAddress+"/"+prefix, "dev", interfaceName)
		if err != nil {
			common.LogError("Failed to set IP address: %v", err)
		}
	} else if common.IsWindows() {
		_, err := common.Execute("netsh", "interface", "ip", "set", "address", interfaceName, "static", ipAddress, prefix)
		if err != nil {
			common.LogError("Failed to set IP address: %v", err)
		}
	} else if common.IsMacOS() {
		_, err := common.Execute("ifconfig", interfaceName, "inet", ipAddress+"/"+prefix)
		if err != nil {
			common.LogError("Failed to set IP address: %v", err)
		}
	} else {
		common.LogError("Setting IP address not implemented for this platform")
	}
}

func setInterfaceState(interfaceName string, up bool) {
	state := "down"
	if up {
		state = "up"
		common.LogInfo("Bringing interface %s up", interfaceName)
	} else {
		common.LogInfo("Bringing interface %s down", interfaceName)
	}

	if common.IsLinux() {
		_, err := common.Execute("ip", "link", "set", interfaceName, state)
		if err != nil {
			common.LogError("Failed to set interface state: %v", err)
		}
	} else if common.IsWindows() {
		action := "disable"
		if up {
			action = "enable"
		}
		_, err := common.Execute("netsh", "interface", "set", "interface", interfaceName, action)
		if err != nil {
			common.LogError("Failed to set interface state: %v", err)
		}
	} else if common.IsMacOS() {
		action := "down"
		if up {
			action = "up"
		}
		_, err := common.Execute("ifconfig", interfaceName, action)
		if err != nil {
			common.LogError("Failed to set interface state: %v", err)
		}
	} else {
		common.LogError("Setting interface state not implemented for this platform")
	}
}

func setInterfaceMTU(interfaceName, mtu string) {
	common.LogInfo("Setting MTU to %s on %s", mtu, interfaceName)

	if common.IsLinux() {
		_, err := common.Execute("ip", "link", "set", interfaceName, "mtu", mtu)
		if err != nil {
			common.LogError("Failed to set MTU: %v", err)
		}
	} else if common.IsWindows() {
		_, err := common.Execute("netsh", "interface", "ipv4", "set", "subinterface", interfaceName, "mtu="+mtu)
		if err != nil {
			common.LogError("Failed to set MTU: %v", err)
		}
	} else if common.IsMacOS() {
		_, err := common.Execute("ifconfig", interfaceName, "mtu", mtu)
		if err != nil {
			common.LogError("Failed to set MTU: %v", err)
		}
	} else {
		common.LogError("Setting MTU not implemented for this platform")
	}
}

func setInterfaceMAC(interfaceName, mac string) {
	common.LogInfo("Setting MAC address to %s on %s", mac, interfaceName)

	if common.IsLinux() {
		_, err := common.Execute("ip", "link", "set", interfaceName, "down")
		if err != nil {
			common.LogError("Failed to bring interface down: %v", err)
			return
		}
		_, err = common.Execute("ip", "link", "set", interfaceName, "address", mac)
		if err != nil {
			common.LogError("Failed to set MAC address: %v", err)
			return
		}
		_, err = common.Execute("ip", "link", "set", interfaceName, "up")
		if err != nil {
			common.LogError("Failed to bring interface up: %v", err)
		}
	} else if common.IsWindows() {
		// On Windows, changing MAC address requires registry changes
		// This is a simplified version that may not work on all systems
		_, err := common.Execute("powershell", "-Command", 
			fmt.Sprintf("Set-NetAdapter -Name '%s' -MacAddress '%s'", interfaceName, strings.ReplaceAll(mac, ":", "")))
		if err != nil {
			common.LogError("Failed to set MAC address: %v", err)
		}
	} else if common.IsMacOS() {
		_, err := common.Execute("ifconfig", interfaceName, "ether", mac)
		if err != nil {
			common.LogError("Failed to set MAC address: %v", err)
		}
	} else {
		common.LogError("Setting MAC address not implemented for this platform")
	}
}
