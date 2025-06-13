package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/RimakiTaema/netmgr/internal/bandwidth"
	"github.com/RimakiTaema/netmgr/internal/common"
	"github.com/RimakiTaema/netmgr/internal/diagnostic"
	"github.com/RimakiTaema/netmgr/internal/dns"
	"github.com/RimakiTaema/netmgr/internal/firewall"
	"github.com/RimakiTaema/netmgr/internal/forward"
	"github.com/RimakiTaema/netmgr/internal/interface_mgmt"
	"github.com/RimakiTaema/netmgr/internal/route"
	"github.com/RimakiTaema/netmgr/internal/tunnel"
)

const (
	version = "1.0.0"
)

func showHelp() {
	fmt.Println(`Network Management Suite - Cross-platform network tool

USAGE:
    netmgr <context> <command> [parameters...]

CONTEXTS:
    interface       Network interface management
    route           Routing table management  
    firewall        Firewall rules management
    forward         Port forwarding management
    dns             DNS configuration
    bandwidth       Traffic shaping and QoS
    tunnel          Tunnel interfaces
    diag            Network diagnostics

GLOBAL OPTIONS:
    -v, --verbose   Enable verbose output
    -n, --dry-run   Show what would be done without executing
    -f, --force     Force operations without confirmation
    -h, --help      Show this help

EXAMPLES:
    netmgr interface show
    netmgr interface set eth0 ip 192.168.1.100 24
    netmgr route add 10.0.0.0/8 via 192.168.1.1
    netmgr firewall rule allow 22 tcp
    netmgr forward add minecraft 25565 10.0.0.2 25565
    netmgr dns set 8.8.8.8 1.1.1.1
    netmgr bandwidth limit eth0 100mbit
    netmgr diag connectivity google.com
    netmgr diag ports 192.168.1.1 22,80,443

For detailed help on specific contexts, use:
    netmgr <context> help`)
}

func parseGlobalOptions(args []string) ([]string, *common.Options) {
	opts := &common.Options{
		Verbose: false,
		DryRun:  false,
		Force:   false,
	}

	var result []string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "-v", "--verbose":
			opts.Verbose = true
		case "-n", "--dry-run":
			opts.DryRun = true
		case "-f", "--force":
			opts.Force = true
		case "-h", "--help":
			showHelp()
			os.Exit(0)
		default:
			if strings.HasPrefix(arg, "-") {
				common.LogError("Unknown option: %s", arg)
				os.Exit(1)
			}
			result = append(result, args[i:]...)
			return result, opts
		}
	}

	return result, opts
}

func main() {
	// Parse global options
	args, opts := parseGlobalOptions(os.Args[1:])
	common.SetOptions(opts)

	// Initialize system
	if err := common.InitSystem(); err != nil {
		common.LogError("Failed to initialize: %v", err)
		os.Exit(1)
	}

	// Check for root privileges on Unix systems
	if !common.IsWindows() && !common.IsRoot() && !opts.DryRun {
		common.LogError("This tool requires administrator privileges")
		os.Exit(1)
	}

	// Check dependencies
	if err := common.CheckDependencies(); err != nil {
		common.LogError("Dependency check failed: %v", err)
		os.Exit(1)
	}

	// Handle commands
	if len(args) < 1 {
		showHelp()
		os.Exit(1)
	}

	context := args[0]
	var command string
	var params []string

	if len(args) > 1 {
		command = args[1]
		params = args[2:]
	} else {
		command = "help"
		params = []string{}
	}

	switch context {
	case "interface", "int":
		interface_mgmt.HandleCommand(command, params)
	case "route", "rt":
		route.HandleCommand(command, params)
	case "firewall", "fw":
		firewall.HandleCommand(command, params)
	case "forward", "fwd":
		forward.HandleCommand(command, params)
	case "dns":
		dns.HandleCommand(command, params)
	case "bandwidth", "bw":
		bandwidth.HandleCommand(command, params)
	case "tunnel", "tun":
		tunnel.HandleCommand(command, params)
	case "diag", "diagnostic":
		diagnostic.HandleCommand(command, params)
	case "version":
		fmt.Printf("Network Management Suite v%s\n", version)
	case "help":
		showHelp()
	default:
		common.LogError("Unknown context: %s", context)
		showHelp()
		os.Exit(1)
	}
}
