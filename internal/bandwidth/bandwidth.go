package bandwidth

import (
	"fmt"
	"strings"

	"github.com/RimakiTaema/netmgr/internal/common"
)

// HandleCommand processes bandwidth management commands
func HandleCommand(command string, params []string) {
	switch command {
	case "show", "list":
		var iface string
		if len(params) > 0 {
			iface = params[0]
		}
		showBandwidth(iface)
	case "limit":
		if len(params) < 2 {
			common.LogError("Usage: bandwidth limit <interface> <rate> [burst]")
			return
		}
		burst := params[1]
		if len(params) > 2 {
			burst = params[2]
		}
		limitBandwidth(params[0], params[1], burst)
	case "help":
		showHelp()
	default:
		common.LogError("Unknown bandwidth command: %s", command)
	}
}

func showHelp() {
	fmt.Println(`Bandwidth Management Commands:

    show [interface]      Show bandwidth configuration for all or specific interface
    limit <interface> <rate> [burst]
                          Set bandwidth limit on interface
                          
Examples:
    netmgr bandwidth show eth0
    netmgr bandwidth limit eth0 100mbit 200mbit`)
}

func showBandwidth(iface string) {
	if iface != "" {
		common.LogInfo("Bandwidth configuration for %s:", iface)
		fmt.Println()
		
		if common.IsLinux() {
			output, err := common.Execute("tc", "qdisc", "show", "dev", iface)
			if err == nil {
				fmt.Println(output)
			}
			
			output, err = common.Execute("tc", "class", "show", "dev", iface)
			if err == nil {
				fmt.Println(output)
			}
		} else if common.IsWindows() {
			output, err := common.Execute("powershell", "-Command", 
				fmt.Sprintf("Get-NetQosPolicy | Where-Object {$_.NetworkProfile -eq '%s'}", iface))
			if err == nil {
				fmt.Println(output)
			}
		} else if common.IsMacOS() {
			output, err := common.Execute("ipfw", "pipe", "show")
			if err == nil {
				lines := strings.Split(output, "\n")
				for _, line := range lines {
					if strings.Contains(line, iface) {
						fmt.Println(line)
					}
				}
			}
		} else {
			common.LogError("Showing bandwidth configuration not implemented for this platform")
		}
	} else {
		common.LogInfo("All interface bandwidth configurations:")
		fmt.Println()
		
		if common.IsLinux() {
			// Get all interfaces
			output, err := common.Execute("ip", "link", "show")
			if err == nil {
				lines := strings.Split(output, "\n")
				var currentIface string
				
				for _, line := range lines {
					if strings.Contains(line, ": ") {
						parts := strings.Split(line, ": ")
						if len(parts) > 1 {
							currentIface = strings.TrimSpace(parts[1])
							currentIface = strings.Split(currentIface, "@")[0]
							
							if currentIface != "lo" {
								fmt.Printf("\n%s=== %s ===%s\n", common.ColorCyan, currentIface, common.ColorReset)
								qdisc, _ := common.Execute("tc", "qdisc", "show", "dev", currentIface)
								fmt.Println(qdisc)
							}
						}
					}
				}
			}
		} else if common.IsWindows() {
			output, err := common.Execute("powershell", "-Command", "Get-NetQosPolicy")
			if err == nil {
				fmt.Println(output)
			}
		} else if common.IsMacOS() {
			output, err := common.Execute("ipfw", "pipe", "show")
			if err == nil {
				fmt.Println(output)
			}
		} else {
			common.LogError("Showing bandwidth configuration not implemented for this platform")
		}
	}
}

func limitBandwidth(iface, rate, burst string) {
	common.LogInfo("Setting bandwidth limit on %s: %s (burst: %s)", iface, rate, burst)
	
	if common.IsLinux() {
		// Remove existing qdisc
		_, _ = common.Execute("tc", "qdisc", "del", "dev", iface, "root")
		
		// Add new rate limiting
		_, err := common.Execute("tc", "qdisc", "add", "dev", iface, "root", "handle", "1:", 
			"tbf", "rate", rate, "burst", burst, "latency", "70ms")
		if err != nil {
			common.LogError("Failed to set bandwidth limit: %v", err)
		}
	} else if common.IsWindows() {
		// Convert rate to bits per second
		rateValue := parseRate(rate)
		
		// Create QoS policy
		policyName := "NetMgr-" + iface
		_, err := common.Execute("powershell", "-Command", 
			fmt.Sprintf("New-NetQosPolicy -Name '%s' -NetworkProfile %s -ThrottleRateActionBitsPerSecond %d", 
				policyName, iface, rateValue))
		if err != nil {
			common.LogError("Failed to set bandwidth limit: %v", err)
		}
	} else if common.IsMacOS() {
		// macOS uses ipfw for traffic shaping
		// This is a simplified version
		rateValue := parseRate(rate)
		
		// Create pipe
		_, err := common.Execute("ipfw", "pipe", "1", "config", "bw", fmt.Sprintf("%dKbit/s", rateValue/1000))
		if err != nil {
			common.LogError("Failed to create pipe: %v", err)
			return
		}
		
		// Assign traffic to pipe
		_, err = common.Execute("ipfw", "add", "100", "pipe", "1", "ip", "from", "any", "to", "any", "via", iface)
		if err != nil {
			common.LogError("Failed to assign traffic to pipe: %v", err)
		}
	} else {
		common.LogError("Setting bandwidth limit not implemented for this platform")
	}
}

// parseRate converts rate strings like "100mbit" to bits per second
func parseRate(rate string) int64 {
	rate = strings.ToLower(rate)
	var multiplier int64 = 1
	
	if strings.HasSuffix(rate, "kbit") || strings.HasSuffix(rate, "kbps") {
		multiplier = 1000
		rate = strings.TrimSuffix(strings.TrimSuffix(rate, "kbps"), "kbit")
	} else if strings.HasSuffix(rate, "mbit") || strings.HasSuffix(rate, "mbps") {
		multiplier = 1000000
		rate = strings.TrimSuffix(strings.TrimSuffix(rate, "mbps"), "mbit")
	} else if strings.HasSuffix(rate, "gbit") || strings.HasSuffix(rate, "gbps") {
		multiplier = 1000000000
		rate = strings.TrimSuffix(strings.TrimSuffix(rate, "gbps"), "gbit")
	}
	
	var value int64
	fmt.Sscanf(rate, "%d", &value)
	return value * multiplier
}
