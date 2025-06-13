package diagnostic

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/yourusername/netmgr/internal/common"
)

// HandleCommand processes diagnostic commands
func HandleCommand(command string, params []string) {
	switch command {
	case "connectivity", "ping":
		target := "8.8.8.8"
		count := "3"
		if len(params) > 0 {
			target = params[0]
		}
		if len(params) > 1 {
			count = params[1]
		}
		testConnectivity(target, count)
	case "ports", "port":
		if len(params) < 1 {
			common.LogError("Usage: diag ports <target> [port_list]")
			return
		}
		ports := "22,80,443,25565"
		if len(params) > 1 {
			ports = params[1]
		}
		testPorts(params[0], ports)
	case "bandwidth", "bw":
		iface := "eth0"
		duration := "10"
		if len(params) > 0 {
			iface = params[0]
		}
		if len(params) > 1 {
			duration = params[1]
		}
		monitorBandwidth(iface, duration)
	case "help":
		showHelp()
	default:
		common.LogError("Unknown diagnostic command: %s", command)
	}
}

func showHelp() {
	fmt.Println(`Network Diagnostic Commands:

    connectivity [target] [count]
                          Test connectivity to target (default: 8.8.8.8)
    ports <target> [port_list]
                          Test if ports are open on target (default ports: 22,80,443,25565)
    bandwidth [interface] [duration]
                          Monitor bandwidth on interface (default: eth0, 10s)

Examples:
    netmgr diag connectivity google.com 5
    netmgr diag ports 192.168.1.1 22,80,443
    netmgr diag bandwidth eth0 30`)
}

func testConnectivity(target, count string) {
	common.LogInfo("Testing connectivity to %s", target)
	fmt.Println()
	
	fmt.Printf("%s=== Ping Test ===%s\n", common.ColorCyan, common.ColorReset)
	
	if common.IsWindows() {
		output, err := common.Execute("ping", "-n", count, target)
		if err == nil {
			fmt.Println(output)
		} else {
			common.LogError("Ping failed: %v", err)
		}
	} else {
		output, err := common.Execute("ping", "-c", count, target)
		if err == nil {
			fmt.Println(output)
		} else {
			common.LogError("Ping failed: %v", err)
		}
	}
	
	fmt.Printf("\n%s=== Traceroute ===%s\n", common.ColorCyan, common.ColorReset)
	
	if common.IsWindows() {
		output, err := common.Execute("tracert", target)
		if err == nil {
			fmt.Println(output)
		} else {
			common.LogError("Traceroute failed: %v", err)
		}
	} else if common.IsMacOS() {
		output, err := common.Execute("traceroute", target)
		if err == nil {
			fmt.Println(output)
		} else {
			common.LogError("Traceroute failed: %v", err)
		}
	} else {
		// Try traceroute first, fall back to tracepath
		output, err := common.Execute("traceroute", target)
		if err != nil {
			output, err = common.Execute("tracepath", target)
			if err != nil {
				common.LogError("Traceroute failed: %v", err)
			}
		}
		fmt.Println(output)
	}
	
	fmt.Printf("\n%s=== DNS Resolution ===%s\n", common.ColorCyan, common.ColorReset)
	
	if common.IsWindows() {
		output, err := common.Execute("nslookup", target)
		if err == nil {
			fmt.Println(output)
		} else {
			common.LogError("DNS resolution failed: %v", err)
		}
	} else {
		output, err := common.Execute("nslookup", target)
		if err == nil {
			fmt.Println(output)
		} else {
			common.LogError("DNS resolution failed: %v", err)
		}
	}
}

func testPorts(target, ports string) {
	common.LogInfo("Testing ports on %s: %s", target, ports)
	fmt.Println()
	
	portList := strings.Split(ports, ",")
	
	for _, port := range portList {
		port = strings.TrimSpace(port)
		
		if common.IsWindows() {
			// Windows uses PowerShell for port testing
			cmd := fmt.Sprintf("$conn = New-Object System.Net.Sockets.TcpClient; try { $conn.Connect('%s', %s); Write-Host 'Open' } catch { Write-Host 'Closed' } finally { $conn.Close() }", target, port)
			output, _ := common.Execute("powershell", "-Command", cmd)
			
			if strings.Contains(output, "Open") {
				fmt.Printf("%s✓%s Port %s is open\n", common.ColorGreen, common.ColorReset, port)
			} else {
				fmt.Printf("%s✗%s Port %s is closed or filtered\n", common.ColorRed, common.ColorReset, port)
			}
		} else {
			// Unix systems use nc (netcat)
			_, err := common.Execute("nc", "-z", "-w", "3", target, port)
			
			if err == nil {
				fmt.Printf("%s✓%s Port %s is open\n", common.ColorGreen, common.ColorReset, port)
			} else {
				fmt.Printf("%s✗%s Port %s is closed or filtered\n", common.ColorRed, common.ColorReset, port)
			}
		}
	}
}

func monitorBandwidth(iface, durationStr string) {
	duration, err := strconv.Atoi(durationStr)
	if err != nil {
		common.LogError("Invalid duration: %s", durationStr)
		return
	}
	
	common.LogInfo("Monitoring bandwidth on %s for %ds", iface, duration)
	fmt.Println()
	
	if common.IsLinux() {
		// Get initial counters
		rxBytesStart, txBytesStart := getLinuxInterfaceCounters(iface)
		
		// Wait for specified duration
		time.Sleep(time.Duration(duration) * time.Second)
		
		// Get final counters
		rxBytesEnd, txBytesEnd := getLinuxInterfaceCounters(iface)
		
		// Calculate rates
		rxDiff := rxBytesEnd - rxBytesStart
		txDiff := txBytesEnd - txBytesStart
		
		rxRate := rxDiff / int64(duration)
		txRate := txDiff / int64(duration)
		
		fmt.Printf("RX: %s/s\n", formatBytes(rxRate))
		fmt.Printf("TX: %s/s\n", formatBytes(txRate))
	} else if common.IsWindows() {
		// Windows uses PowerShell for bandwidth monitoring
		cmd := fmt.Sprintf(`
$adapter = Get-NetAdapter | Where-Object {$_.Name -eq '%s' -or $_.InterfaceDescription -like '*%s*'} | Select-Object -First 1
$startStats = $adapter | Get-NetAdapterStatistics
Start-Sleep -Seconds %d
$endStats = $adapter | Get-NetAdapterStatistics
$rxDiff = $endStats.ReceivedBytes - $startStats.ReceivedBytes
$txDiff = $endStats.SentBytes - $startStats.SentBytes
$rxRate = $rxDiff / %d
$txRate = $txDiff / %d
"RX: " + [math]::Round($rxRate / 1KB, 2) + " KB/s"
"TX: " + [math]::Round($txRate / 1KB, 2) + " KB/s"
`, iface, iface, duration, duration, duration)
		
		output, err := common.Execute("powershell", "-Command", cmd)
		if err == nil {
			fmt.Println(output)
		} else {
			common.LogError("Failed to monitor bandwidth: %v", err)
		}
	} else if common.IsMacOS() {
		// macOS uses netstat for bandwidth monitoring
		cmd := fmt.Sprintf("netstat -I %s -b -w %d 2", iface, duration)
		output, err := common.Execute("sh", "-c", cmd)
		if err == nil {
			fmt.Println(output)
		} else {
			common.LogError("Failed to monitor bandwidth: %v", err)
		}
	} else {
		common.LogError("Bandwidth monitoring not implemented for this platform")
	}
}

func getLinuxInterfaceCounters(iface string) (int64, int64) {
	var rxBytes, txBytes int64
	
	rxData, err := common.Execute("cat", fmt.Sprintf("/sys/class/net/%s/statistics/rx_bytes", iface))
	if err == nil {
		rxBytes, _ = strconv.ParseInt(strings.TrimSpace(rxData), 10, 64)
	}
	
	txData, err := common.Execute("cat", fmt.Sprintf("/sys/class/net/%s/statistics/tx_bytes", iface))
	if err == nil {
		txBytes, _ = strconv.ParseInt(strings.TrimSpace(txData), 10, 64)
	}
	
	return rxBytes, txBytes
}

func formatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)
	
	if bytes < KB {
		return fmt.Sprintf("%d B", bytes)
	} else if bytes < MB {
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	} else if bytes < GB {
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	} else {
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	}
}
