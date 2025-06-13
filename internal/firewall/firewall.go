package firewall

import (
	"fmt"
	"strings"

	"github.com/yourusername/netmgr/internal/common"
)

// HandleCommand processes firewall commands
func HandleCommand(command string, params []string) {
	switch command {
	case "show", "list":
		var table string
		if len(params) > 0 {
			table = params[0]
		}
		showFirewall(table)
	case "rule":
		if len(params) < 1 {
			common.LogError("Usage: firewall rule <action> [parameters...]")
			return
		}
		handleFirewallRule(params[0], params[1:])
	case "help":
		showHelp()
	default:
		common.LogError("Unknown firewall command: %s", command)
	}
}

func showHelp() {
	fmt.Println(`Firewall Management Commands:

    show [table]          Show firewall rules (default: filter)
                          Tables: filter, nat, mangle, raw
    rule <action> [parameters...]
                          Manage firewall rules
                          
Actions:
    allow <port> [protocol] [interface]
                          Allow traffic on specified port
    deny <port> [protocol] [interface]
                          Block traffic on specified port
    flush [table]         Clear all rules in table
    save [file]           Save rules to file
    restore [file]        Restore rules from file

Examples:
    netmgr firewall show
    netmgr firewall show nat
    netmgr firewall rule allow 22 tcp
    netmgr firewall rule deny 25 tcp
    netmgr firewall rule flush`)
}

func showFirewall(table string) {
	if table == "" {
		table = "filter"
	}

	common.LogInfo("Firewall rules (table: %s):", table)
	fmt.Println()

	if common.IsLinux() {
		args := []string{"-t", table, "-L", "-n", "-v", "--line-numbers"}
		output, err := common.Execute("iptables", args...)
		if err == nil {
			fmt.Println(output)
		}
	} else if common.IsWindows() {
		output, err := common.Execute("netsh", "advfirewall", "firewall", "show", "rule", "name=all")
		if err == nil {
			fmt.Println(output)
		}
	} else if common.IsMacOS() {
		output, err := common.Execute("pfctl", "-s", "rules")
		if err == nil {
			fmt.Println(output)
		}
	} else {
		common.LogError("Showing firewall rules not implemented for this platform")
	}
}

func handleFirewallRule(action string, params []string) {
	switch action {
	case "allow":
		if len(params) < 1 {
			common.LogError("Usage: firewall rule allow <port> [protocol] [interface]")
			return
		}
		port := params[0]
		protocol := "tcp"
		var iface string
		
		if len(params) > 1 {
			protocol = params[1]
		}
		if len(params) > 2 {
			iface = params[2]
		}
		
		allowPort(port, protocol, iface)
	case "deny", "block":
		if len(params) < 1 {
			common.LogError("Usage: firewall rule deny <port> [protocol] [interface]")
			return
		}
		port := params[0]
		protocol := "tcp"
		var iface string
		
		if len(params) > 1 {
			protocol = params[1]
		}
		if len(params) > 2 {
			iface = params[2]
		}
		
		denyPort(port, protocol, iface)
	case "flush":
		var table string
		if len(params) > 0 {
			table = params[0]
		} else {
			table = "filter"
		}
		flushFirewall(table)
	case "save":
		var file string
		if len(params) > 0 {
			file = params[0]
		} else {
			if common.IsWindows() {
				file = "C:\\Windows\\System32\\firewall_rules.wfw"
			} else {
				file = "/etc/iptables/rules.v4"
			}
		}
		saveFirewall(file)
	case "restore":
		var file string
		if len(params) > 0 {
			file = params[0]
		} else {
			if common.IsWindows() {
				file = "C:\\Windows\\System32\\firewall_rules.wfw"
			} else {
				file = "/etc/iptables/rules.v4"
			}
		}
		restoreFirewall(file)
	default:
		common.LogError("Unknown firewall rule action: %s", action)
	}
}

func allowPort(port, protocol, iface string) {
	common.LogInfo("Adding allow rule for port %s/%s", port, protocol)
	
	if common.IsLinux() {
		args := []string{"-A", "INPUT", "-p", protocol, "--dport", port, "-j", "ACCEPT"}
		if iface != "" {
			args = []string{"-A", "INPUT", "-i", iface, "-p", protocol, "--dport", port, "-j", "ACCEPT"}
		}
		
		_, err := common.Execute("iptables", args...)
		if err != nil {
			common.LogError("Failed to add allow rule: %v", err)
		}
	} else if common.IsWindows() {
		name := fmt.Sprintf("NetMgr-Allow-%s-%s", protocol, port)
		args := []string{"advfirewall", "firewall", "add", "rule", 
			fmt.Sprintf("name=%s", name),
			fmt.Sprintf("protocol=%s", protocol),
			fmt.Sprintf("localport=%s", port),
			"dir=in", "action=allow"}
		
		if iface != "" {
			args = append(args, fmt.Sprintf("interface=%s", iface))
		}
		
		_, err := common.Execute("netsh", args...)
		if err != nil {
			common.LogError("Failed to add allow rule: %v", err)
		}
	} else if common.IsMacOS() {
		// This is a simplified version and may need root privileges
		rule := fmt.Sprintf("pass in proto %s from any to any port %s", protocol, port)
		if iface != "" {
			rule = fmt.Sprintf("pass in on %s proto %s from any to any port %s", iface, protocol, port)
		}
		
		_, err := common.Execute("echo", rule, "|", "pfctl", "-a", "com.netmgr/rules", "-f", "-")
		if err != nil {
			common.LogError("Failed to add allow rule: %v", err)
		}
	} else {
		common.LogError("Adding firewall rules not implemented for this platform")
	}
}

func denyPort(port, protocol, iface string) {
	common.LogInfo("Adding deny rule for port %s/%s", port, protocol)
	
	if common.IsLinux() {
		args := []string{"-A", "INPUT", "-p", protocol, "--dport", port, "-j", "DROP"}
		if iface != "" {
			args = []string{"-A", "INPUT", "-i", iface, "-p", protocol, "--dport", port, "-j", "DROP"}
		}
		
		_, err := common.Execute("iptables", args...)
		if err != nil {
			common.LogError("Failed to add deny rule: %v", err)
		}
	} else if common.IsWindows() {
		name := fmt.Sprintf("NetMgr-Block-%s-%s", protocol, port)
		args := []string{"advfirewall", "firewall", "add", "rule", 
			fmt.Sprintf("name=%s", name),
			fmt.Sprintf("protocol=%s", protocol),
			fmt.Sprintf("localport=%s", port),
			"dir=in", "action=block"}
		
		if iface != "" {
			args = append(args, fmt.Sprintf("interface=%s", iface))
		}
		
		_, err := common.Execute("netsh", args...)
		if err != nil {
			common.LogError("Failed to add deny rule: %v", err)
		}
	} else if common.IsMacOS() {
		// This is a simplified version and may need root privileges
		rule := fmt.Sprintf("block in proto %s from any to any port %s", protocol, port)
		if iface != "" {
			rule = fmt.Sprintf("block in on %s proto %s from any to any port %s", iface, protocol, port)
		}
		
		_, err := common.Execute("echo", rule, "|", "pfctl", "-a", "com.netmgr/rules", "-f", "-")
		if err != nil {
			common.LogError("Failed to add deny rule: %v", err)
		}
	} else {
		common.LogError("Adding firewall rules not implemented for this platform")
	}
}

func flushFirewall(table string) {
	common.LogInfo("Flushing %s table", table)
	
	if common.IsLinux() {
		_, err := common.Execute("iptables", "-t", table, "-F")
		if err != nil {
			common.LogError("Failed to flush firewall: %v", err)
		}
	} else if common.IsWindows() {
		_, err := common.Execute("netsh", "advfirewall", "reset")
		if err != nil {
			common.LogError("Failed to reset firewall: %v", err)
		}
	} else if common.IsMacOS() {
		_, err := common.Execute("pfctl", "-F", "rules")
		if err != nil {
			common.LogError("Failed to flush firewall: %v", err)
		}
	} else {
		common.LogError("Flushing firewall not implemented for this platform")
	}
}

func saveFirewall(file string) {
	common.LogInfo("Saving firewall rules to %s", file)
	
	if common.IsLinux() {
		_, err := common.Execute("sh", "-c", fmt.Sprintf("iptables-save > %s", file))
		if err != nil {
			common.LogError("Failed to save firewall rules: %v", err)
		}
	} else if common.IsWindows() {
		_, err := common.Execute("netsh", "advfirewall", "export", file)
		if err != nil {
			common.LogError("Failed to export firewall rules: %v", err)
		}
	} else if common.IsMacOS() {
		_, err := common.Execute("sh", "-c", fmt.Sprintf("pfctl -sr > %s", file))
		if err != nil {
			common.LogError("Failed to save firewall rules: %v", err)
		}
	} else {
		common.LogError("Saving firewall rules not implemented for this platform")
	}
}

func restoreFirewall(file string) {
	common.LogInfo("Restoring firewall rules from %s", file)
	
	if common.IsLinux() {
		_, err := common.Execute("sh", "-c", fmt.Sprintf("iptables-restore < %s", file))
		if err != nil {
			common.LogError("Failed to restore firewall rules: %v", err)
		}
	} else if common.IsWindows() {
		_, err := common.Execute("netsh", "advfirewall", "import", file)
		if err != nil {
			common.LogError("Failed to import firewall rules: %v", err)
		}
	} else if common.IsMacOS() {
		_, err := common.Execute("sh", "-c", fmt.Sprintf("pfctl -f %s", file))
		if err != nil {
			common.LogError("Failed to restore firewall rules: %v", err)
		}
	} else {
		common.LogError("Restoring firewall rules not implemented for this platform")
	}
}
