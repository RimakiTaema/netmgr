package tunnel

import (
	"fmt"

	"github.com/RimakiTaema/netmgr/internal/common"
)

// HandleCommand processes tunnel commands
func HandleCommand(command string, params []string) {
	switch command {
	case "create", "add":
		if len(params) < 4 {
			common.LogError("Usage: tunnel create <name> <type> <local_ip> <remote_ip>")
			return
		}
		createTunnel(params[0], params[1], params[2], params[3])
	case "delete", "del":
		if len(params) < 1 {
			common.LogError("Usage: tunnel delete <name>")
			return
		}
		deleteTunnel(params[0])
	case "help":
		showHelp()
	default:
		common.LogError("Unknown tunnel command: %s", command)
	}
}

func showHelp() {
	fmt.Println(`Tunnel Management Commands:

    create <name> <type> <local_ip> <remote_ip>
                          Create a new tunnel
    delete <name>         Delete a tunnel

Tunnel types:
    gre                   Generic Routing Encapsulation
    ipip                  IP-in-IP encapsulation
    sit                   Simple Internet Transition (IPv6-over-IPv4)

Examples:
    netmgr tunnel create mytunnel gre 192.168.1.1 10.0.0.1
    netmgr tunnel delete mytunnel`)
}

func createTunnel(name, tunnelType, localIP, remoteIP string) {
	common.LogInfo("Creating %s tunnel: %s", tunnelType, name)
	
	if common.IsLinux() {
		var mode string
		switch tunnelType {
		case "gre":
			mode = "gre"
		case "ipip":
			mode = "ipip"
		case "sit":
			mode = "sit"
		default:
			common.LogError("Unsupported tunnel type: %s", tunnelType)
			return
		}
		
		_, err := common.Execute("ip", "tunnel", "add", name, "mode", mode, 
			"remote", remoteIP, "local", localIP)
		if err != nil {
			common.LogError("Failed to create tunnel: %v", err)
			return
		}
		
		_, err = common.Execute("ip", "link", "set", name, "up")
		if err != nil {
			common.LogError("Failed to bring tunnel up: %v", err)
		}
	} else if common.IsWindows() {
		// Windows uses different commands for tunnels
		switch tunnelType {
		case "gre":
			_, err := common.Execute("netsh", "interface", "ipv4", "add", "interface", 
				name, "type=tunnel", fmt.Sprintf("source=%s", localIP), 
				fmt.Sprintf("destination=%s", remoteIP))
			if err != nil {
				common.LogError("Failed to create GRE tunnel: %v", err)
				return
			}
		case "ipip", "sit":
			common.LogError("Tunnel type %s not supported on Windows", tunnelType)
			return
		default:
			common.LogError("Unsupported tunnel type: %s", tunnelType)
			return
		}
	} else {
		common.LogError("Creating tunnels not implemented for this platform")
	}
}

func deleteTunnel(name string) {
	common.LogInfo("Deleting tunnel: %s", name)
	
	if common.IsLinux() {
		_, err := common.Execute("ip", "link", "set", name, "down")
		if err != nil {
			common.LogError("Failed to bring tunnel down: %v", err)
		}
		
		_, err = common.Execute("ip", "tunnel", "del", name)
		if err != nil {
			common.LogError("Failed to delete tunnel: %v", err)
		}
	} else if common.IsWindows() {
		_, err := common.Execute("netsh", "interface", "ipv4", "delete", "interface", name)
		if err != nil {
			common.LogError("Failed to delete tunnel: %v", err)
		}
	} else {
		common.LogError("Deleting tunnels not implemented for this platform")
	}
}
