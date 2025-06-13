package forward

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/RimakiTaema/netmgr/internal/common"
)

// PortForward represents a port forwarding rule
type PortForward struct {
	Name      string    `json:"name"`
	SrcPort   string    `json:"src_port"`
	DestIP    string    `json:"dest_ip"`
	DestPort  string    `json:"dest_port"`
	Protocol  string    `json:"protocol"`
	Created   time.Time `json:"created"`
	Active    bool      `json:"active"`
}

// ForwardConfig is a map of port forwards by name
type ForwardConfig map[string]PortForward

// HandleCommand processes port forwarding commands
func HandleCommand(command string, params []string) {
	switch command {
	case "show", "list":
		showForwards()
	case "add":
		if len(params) < 4 {
			common.LogError("Usage: forward add <name> <src_port> <dest_ip> <dest_port> [protocol]")
			return
		}
		protocol := "tcp"
		if len(params) > 4 {
			protocol = params[4]
		}
		addForward(params[0], params[1], params[2], params[3], protocol)
	case "remove", "del":
		if len(params) < 1 {
			common.LogError("Usage: forward remove <name>")
			return
		}
		removeForward(params[0])
	case "help":
		showHelp()
	default:
		common.LogError("Unknown forward command: %s", command)
	}
}

func showHelp() {
	fmt.Println(`Port Forwarding Commands:

    show                  Show active port forwards
    add <name> <src_port> <dest_ip> <dest_port> [protocol]
                          Add a port forward
    remove <name>         Remove a port forward

Examples:
    netmgr forward show
    netmgr forward add web 80 192.168.1.100 8080 tcp
    netmgr forward remove web`)
}

func showForwards() {
	common.LogInfo("Active port forwards:")
	fmt.Println()

	// Header
	fmt.Printf("%-15s %-10s %-25s %-10s %-20s\n", "NAME", "PROTOCOL", "FORWARD", "STATUS", "CREATED")
	fmt.Printf("%-15s %-10s %-25s %-10s %-20s\n", "----", "--------", "-------", "------", "-------")

	// Load config
	config := loadForwardConfig()
	
	// Display forwards
	for name, forward := range config {
		status := "INACTIVE"
		if forward.Active {
			status = "ACTIVE"
		}
		
		forwardStr := fmt.Sprintf("%s->%s:%s", forward.SrcPort, forward.DestIP, forward.DestPort)
		created := forward.Created.Format("2006-01-02 15:04:05")
		
		fmt.Printf("%-15s %-10s %-25s %-10s %-20s\n", name, forward.Protocol, forwardStr, status, created)
	}
}

func addForward(name, srcPort, destIP, destPort, protocol string) {
	// Load config
	config := loadForwardConfig()
	
	// Check if name already exists
	if _, exists := config[name]; exists {
		common.LogError("Port forward with name '%s' already exists", name)
		return
	}
	
	common.LogInfo("Adding port forward: %s (%s -> %s:%s)", name, srcPort, destIP, destPort)
	
	// Enable IP forwarding
	enableIPForwarding()
	
	// Add forwarding rules based on platform
	success := false
	
	if common.IsLinux() {
		success = addLinuxForward(name, srcPort, destIP, destPort, protocol)
	} else if common.IsWindows() {
		success = addWindowsForward(name, srcPort, destIP, destPort, protocol)
	} else if common.IsMacOS() {
		success = addMacOSForward(name, srcPort, destIP, destPort, protocol)
	} else {
		common.LogError("Port forwarding not implemented for this platform")
		return
	}
	
	if success {
		// Save to config
		forward := PortForward{
			Name:      name,
			SrcPort:   srcPort,
			DestIP:    destIP,
			DestPort:  destPort,
			Protocol:  protocol,
			Created:   time.Now(),
			Active:    true,
		}
		
		config[name] = forward
		saveForwardConfig(config)
	}
}

func removeForward(name string) {
	// Load config
	config := loadForwardConfig()
	
	// Check if name exists
	forward, exists := config[name]
	if !exists {
		common.LogError("Port forward with name '%s' does not exist", name)
		return
	}
	
	common.LogInfo("Removing port forward: %s", name)
	
	// Remove forwarding rules based on platform
	if common.IsLinux() {
		removeLinuxForward(name, forward.SrcPort, forward.DestIP, forward.DestPort, forward.Protocol)
	} else if common.IsWindows() {
		removeWindowsForward(name, forward.SrcPort, forward.DestIP, forward.DestPort, forward.Protocol)
	} else if common.IsMacOS() {
		removeMacOSForward(name, forward.SrcPort, forward.DestIP, forward.DestPort, forward.Protocol)
	} else {
		common.LogError("Port forwarding not implemented for this platform")
		return
	}
	
	// Remove from config
	delete(config, name)
	saveForwardConfig(config)
}

func loadForwardConfig() ForwardConfig {
	config := make(ForwardConfig)
	
	configPath := filepath.Join(common.ConfigDir, "forwarding.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		if !os.IsNotExist(err) {
			common.LogError("Failed to read forwarding config: %v", err)
		}
		return config
	}
	
	if err := json.Unmarshal(data, &config); err != nil {
		common.LogError("Failed to parse forwarding config: %v", err)
		return make(ForwardConfig)
	}
	
	return config
}

func saveForwardConfig(config ForwardConfig) {
	configPath := filepath.Join(common.ConfigDir, "forwarding.json")
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		common.LogError("Failed to marshal forwarding config: %v", err)
		return
	}
	
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		common.LogError("Failed to write forwarding config: %v", err)
	}
}

func enableIPForwarding() {
	if common.IsLinux() {
		_, err := common.Execute("sysctl", "-w", "net.ipv4.ip_forward=1")
		if err != nil {
			common.LogError("Failed to enable IP forwarding: %v", err)
		}
	} else if common.IsWindows() {
		// Windows enables forwarding per interface
		// This is handled in the addWindowsForward function
	} else if common.IsMacOS() {
		_, err := common.Execute("sysctl", "-w", "net.inet.ip.forwarding=1")
		if err != nil {
			common.LogError("Failed to enable IP forwarding: %v", err)
		}
	}
}

func addLinuxForward(name, srcPort, destIP, destPort, protocol string) bool {
	// Add DNAT rule
	_, err := common.Execute("iptables", "-t", "nat", "-A", "PREROUTING", 
		"-p", protocol, "--dport", srcPort, 
		"-j", "DNAT", "--to-destination", destIP+":"+destPort,
		"-m", "comment", "--comment", "NETMGR:"+name)
	if err != nil {
		common.LogError("Failed to add DNAT rule: %v", err)
		return false
	}
	
	// Add FORWARD rule
	_, err = common.Execute("iptables", "-A", "FORWARD", 
		"-p", protocol, "-d", destIP, "--dport", destPort,
		"-j", "ACCEPT",
		"-m", "comment", "--comment", "NETMGR:"+name)
	if err != nil {
		common.LogError("Failed to add FORWARD rule: %v", err)
		return false
	}
	
	// Add MASQUERADE rule
	_, err = common.Execute("iptables", "-t", "nat", "-A", "POSTROUTING", 
		"-p", protocol, "-d", destIP, "--dport", destPort,
		"-j", "MASQUERADE",
		"-m", "comment", "--comment", "NETMGR:"+name)
	if err != nil {
		common.LogError("Failed to add MASQUERADE rule: %v", err)
		return false
	}
	
	return true
}

func removeLinuxForward(name, srcPort, destIP, destPort, protocol string) {
	// Use grep to find and remove rules with the specific comment
	_, err := common.Execute("sh", "-c", 
		fmt.Sprintf("iptables-save | grep -v 'NETMGR:%s' | iptables-restore", name))
	if err != nil {
		common.LogError("Failed to remove port forward rules: %v", err)
	}
}

func addWindowsForward(name, srcPort, destIP, destPort, protocol string) bool {
	// Windows uses netsh portproxy
	_, err := common.Execute("netsh", "interface", "portproxy", "add", "v4tov4", 
		"listenport="+srcPort, "listenaddress=0.0.0.0",
		"connectport="+destPort, "connectaddress="+destIP,
		"protocol="+protocol)
	if err != nil {
		common.LogError("Failed to add port forward: %v", err)
		return false
	}
	
	// Add firewall rule to allow incoming traffic
	ruleName := "NetMgr-Forward-" + name
	_, err = common.Execute("netsh", "advfirewall", "firewall", "add", "rule",
		"name="+ruleName, "dir=in", "action=allow",
		"protocol="+protocol, "localport="+srcPort)
	if err != nil {
		common.LogError("Failed to add firewall rule: %v", err)
		return false
	}
	
	return true
}

func removeWindowsForward(name, srcPort, destIP, destPort, protocol string) {
	// Remove portproxy rule
	_, err := common.Execute("netsh", "interface", "portproxy", "delete", "v4tov4",
		"listenport="+srcPort, "listenaddress=0.0.0.0",
		"protocol="+protocol)
	if err != nil {
		common.LogError("Failed to remove port forward: %v", err)
	}
	
	// Remove firewall rule
	ruleName := "NetMgr-Forward-" + name
	_, err = common.Execute("netsh", "advfirewall", "firewall", "delete", "rule",
		"name="+ruleName)
	if err != nil {
		common.LogError("Failed to remove firewall rule: %v", err)
	}
}

func addMacOSForward(name, srcPort, destIP, destPort, protocol string) bool {
	// macOS uses pfctl for port forwarding
	// Create a temporary pf.conf file
	ruleFile := filepath.Join(os.TempDir(), "netmgr_pf_"+name+".conf")
	rule := fmt.Sprintf("rdr pass on lo0 proto %s from any to any port %s -> %s port %s\n", 
		protocol, srcPort, destIP, destPort)
	
	err := os.WriteFile(ruleFile, []byte(rule), 0644)
	if err != nil {
		common.LogError("Failed to create pf rule file: %v", err)
		return false
	}
	
	// Load the rule
	_, err = common.Execute("pfctl", "-a", "com.netmgr/"+name, "-f", ruleFile)
	if err != nil {
		common.LogError("Failed to load pf rule: %v", err)
		os.Remove(ruleFile)
		return false
	}
	
	// Enable pf if not already enabled
	_, err = common.Execute("pfctl", "-e")
	if err != nil {
		common.LogWarn("Failed to enable pf (might already be enabled): %v", err)
	}
	
	// Clean up
	os.Remove(ruleFile)
	return true
}

func removeMacOSForward(name, srcPort, destIP, destPort, protocol string) {
	// Remove the anchor
	_, err := common.Execute("pfctl", "-a", "com.netmgr/"+name, "-F", "all")
	if err != nil {
		common.LogError("Failed to remove pf rule: %v", err)
	}
}
