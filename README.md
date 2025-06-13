# Network Management Suite

A cross-platform network management tool inspired by Windows' netsh, but designed to work on Linux, macOS, and Windows.

## Features

- Interface management (show, configure interfaces)
- Routing management (show, add, delete routes)
- Firewall management (show, add rules)
- Port forwarding management
- DNS configuration
- Bandwidth management
- Tunnel creation and management
- Network diagnostics

## ⚠️ Important Warnings

🛑 **If you are running on Termux proot or emulated Linux via phone:**
I don't guarantee it will work since it utilizes iptables and other system-level commands that require root access. Many network operations may fail or behave unexpectedly in containerized or emulated environments.

⚠️ **Root/Administrator privileges required:**
This tool modifies system network settings and requires elevated privileges to function properly.

## Installation

### From Source

1. Clone the repository:
   \`\`\`
   git clone https://github.com/yourusername/netmgr.git
   cd netmgr
   \`\`\`

2. Build for your platform:
   \`\`\`
   make build
   \`\`\`

3. Install (Linux/macOS):
   \`\`\`
   make install
   \`\`\`

### Cross-Compilation

To build for multiple platforms:

\`\`\`
make build-all
\`\`\`

This will create binaries for various platforms in the `build` directory.

## Usage

\`\`\`
netmgr <context> <command> [parameters...]
\`\`\`

### Contexts

- `interface`: Network interface management
- `route`: Routing table management
- `firewall`: Firewall rules management
- `forward`: Port forwarding management
- `dns`: DNS configuration
- `bandwidth`: Traffic shaping and QoS
- `tunnel`: Tunnel interfaces
- `diag`: Network diagnostics

### Global Options

- `-v, --verbose`: Enable verbose output
- `-n, --dry-run`: Show what would be done without executing
- `-f, --force`: Force operations without confirmation
- `-h, --help`: Show help

### Examples

\`\`\`
# Show all interfaces
netmgr interface show

# Set IP address on interface
netmgr interface set eth0 ip 192.168.1.100 24

# Add a route
netmgr route add 10.0.0.0/8 via 192.168.1.1

# Allow SSH traffic
netmgr firewall rule allow 22 tcp

# Add port forwarding
netmgr forward add minecraft 25565 10.0.0.2 25565

# Set DNS servers
netmgr dns set 8.8.8.8 1.1.1.1

# Limit bandwidth
netmgr bandwidth limit eth0 100mbit

# Test connectivity
netmgr diag connectivity google.com
\`\`\`

## Platform Support

The tool adapts its commands based on the operating system:

- **Linux**: Uses `ip`, `iptables`, `tc`, etc.
- **Windows**: Uses `netsh`, `route`, PowerShell commands
- **macOS**: Uses `networksetup`, `pfctl`, `ipfw`, etc.

Some features may have limited functionality on certain platforms.

## Troubleshooting

### Common Issues

- **Permission denied errors**: Make sure you're running the tool with administrator/root privileges
- **Command not found errors**: Ensure required system utilities are installed
- **Network interface not found**: Verify the interface name is correct for your system

### Platform-Specific Notes

- **Windows**: Some features require PowerShell and may prompt for UAC elevation
- **macOS**: System Integrity Protection may block certain operations
- **Linux**: Different distributions may have different paths for network utilities

## License

MIT
\`\`\`

I've added the warning with a red stop sign emoji to make it highly visible, and I've also expanded the README with a troubleshooting section that might be helpful for users. The warning clearly states that the tool may not work properly in Termux proot or emulated Linux environments on phones due to the need for root access and system-level networking commands.

Would you like me to make any other additions or changes to the documentation?
