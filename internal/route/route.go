package route

import (
	"fmt"
	"strings"

	"github.com/yourusername/netmgr/internal/common"
)

// HandleCommand processes routing commands
func HandleCommand(command string, params []string) {
	switch command {
	case "show", "list":
		var routeType string
		if len(params) > 0 {
			routeType = params[0]
		}
		showRoutes(routeType)
	case "add":
		if len(params) < 1 {
			common.LogError("Usage: route add <destination> [via <gateway>] [dev <interface>] [table <table>]")
			return
		}
		addRoute(params)
	case "delete", "del":
		if len(params) < 1 {
			common.LogError("Usage: route delete <destination> [table <table>]")
			return
		}
		deleteRoute(params)
	case "help":
		showHelp()
	default:
		common.LogError("Unknown route command: %s", command)
	}
}

func showHelp() {
	fmt.Println(`Route Management Commands:

    show [type]           Show routing table (default: main)
                          Types: table, cache, all, or specific route
    add <destination> [via <gateway>] [dev <interface>] [table <table>]
                          Add a route
    delete <destination> [table <table>]
                          Delete a route

Examples:
    netmgr route show
    netmgr route show all
    netmgr route add 10.0.0.0/8 via 192.168.1.1
    netmgr route add 10.0.0.0/8 dev eth0
    netmgr route delete 10.0.0.0/8`)
}

func showRoutes(routeType string) {
	if routeType == "" {
		routeType = "table"
	}

	switch routeType {
	case "table":
		common.LogInfo("Routing table:")
		fmt.Println()
		if common.IsLinux() {
			output, err := common.Execute("ip", "route", "show", "table", "main")
			if err == nil {
				fmt.Println(output)
			}
		} else if common.IsWindows() {
			output, err := common.Execute("route", "print", "-4")
			if err == nil {
				fmt.Println(output)
			}
		} else if common.IsMacOS() {
			output, err := common.Execute("netstat", "-nr", "-f", "inet")
			if err == nil {
				fmt.Println(output)
			}
		} else {
			common.LogError("Showing routes not implemented for this platform")
		}
	case "cache":
		common.LogInfo("Route cache:")
		fmt.Println()
		if common.IsLinux() {
			output, err := common.Execute("ip", "route", "show", "cache")
			if err == nil {
				fmt.Println(output)
			}
		} else {
			common.LogInfo("Route cache display not supported on this platform")
		}
	case "all":
		common.LogInfo("All routing tables:")
		fmt.Println()
		if common.IsLinux() {
			output, err := common.Execute("ip", "route", "show", "table", "all")
			if err == nil {
				tables := make(map[string]bool)
				lines := strings.Split(output, "\n")
				
				// Extract table names
				for _, line := range lines {
					if strings.Contains(line, "table") {
						parts := strings.Split(line, "table")
						if len(parts) > 1 {
							tableName := strings.TrimSpace(parts[1])
							tables[tableName] = true
						}
					}
				}
				
				// Show routes for each table
				for table := range tables {
					fmt.Printf("\n%s=== Table: %s ===%s\n", common.ColorCyan, table, common.ColorReset)
					tableOutput, err := common.Execute("ip", "route", "show", "table", table)
					if err == nil {
						fmt.Println(tableOutput)
					}
				}
			}
		} else if common.IsWindows() {
			output, err := common.Execute("route", "print")
			if err == nil {
				fmt.Println(output)
			}
		} else if common.IsMacOS() {
			output, err := common.Execute("netstat", "-nr")
			if err == nil {
				fmt.Println(output)
			}
		} else {
			common.LogError("Showing all routes not implemented for this platform")
		}
	default:
		common.LogInfo("Routes for %s:", routeType)
		fmt.Println()
		if common.IsLinux() {
			output, err := common.Execute("ip", "route", "show", routeType)
			if err == nil {
				fmt.Println(output)
			}
		} else {
			common.LogError("Showing specific routes not implemented for this platform")
		}
	}
}

func addRoute(params []string) {
	dest := params[0]
	var via, dev, table string
	
	// Parse parameters
	for i := 1; i < len(params); i++ {
		switch params[i] {
		case "via":
			if i+1 < len(params) {
				via = params[i+1]
				i++
			}
		case "dev":
			if i+1 < len(params) {
				dev = params[i+1]
				i++
			}
		case "table":
			if i+1 < len(params) {
				table = params[i+1]
				i++
			}
		}
	}
	
	if table == "" {
		table = "main"
	}
	
	common.LogInfo("Adding route: %s", dest)
	
	if common.IsLinux() {
		args := []string{"route", "add", dest}
		if via != "" {
			args = append(args, "via", via)
		}
		if dev != "" {
			args = append(args, "dev", dev)
		}
		if table != "main" {
			args = append(args, "table", table)
		}
		
		_, err := common.Execute("ip", args...)
		if err != nil {
			common.LogError("Failed to add route: %v", err)
		}
	} else if common.IsWindows() {
		args := []string{"add", dest}
		if via != "" {
			args = append(args, "mask", "255.255.255.0", via)
		}
		if dev != "" {
			args = append(args, "if", dev)
		}
		
		_, err := common.Execute("route", args...)
		if err != nil {
			common.LogError("Failed to add route: %v", err)
		}
	} else if common.IsMacOS() {
		args := []string{"add", "-net", dest}
		if via != "" {
			args = append(args, via)
		}
		
		_, err := common.Execute("route", args...)
		if err != nil {
			common.LogError("Failed to add route: %v", err)
		}
	} else {
		common.LogError("Adding routes not implemented for this platform")
	}
}

func deleteRoute(params []string) {
	dest := params[0]
	var table string
	
	if len(params) > 1 && params[1] == "table" && len(params) > 2 {
		table = params[2]
	} else {
		table = "main"
	}
	
	common.LogInfo("Deleting route: %s", dest)
	
	if common.IsLinux() {
		args := []string{"route", "del", dest}
		if table != "main" {
			args = append(args, "table", table)
		}
		
		_, err := common.Execute("ip", args...)
		if err != nil {
			common.LogError("Failed to delete route: %v", err)
		}
	} else if common.IsWindows() {
		_, err := common.Execute("route", "delete", dest)
		if err != nil {
			common.LogError("Failed to delete route: %v", err)
		}
	} else if common.IsMacOS() {
		_, err := common.Execute("route", "delete", dest)
		if err != nil {
			common.LogError("Failed to delete route: %v", err)
		}
	} else {
		common.LogError("Deleting routes not implemented for this platform")
	}
}
