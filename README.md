# NetMgr - Cross-Platform Network Management Tool

A powerful, cross-platform network management tool written in Rust, inspired by Windows' netsh but designed to work seamlessly on Linux, macOS, and Windows.

## 🚀 Features

- **Interface Management** - Show, configure, and manage network interfaces
- **Routing Management** - View and modify routing tables
- **Firewall Management** - Configure firewall rules across platforms
- **Port Forwarding** - Set up and manage port forwards
- **DNS Configuration** - Manage DNS settings
- **Bandwidth Management** - Traffic shaping and QoS
- **Tunnel Management** - Create and manage network tunnels
- **Network Diagnostics** - Connectivity tests, port scanning, bandwidth monitoring

## ⚠️ Important Warnings

🛑 **Root/Administrator privileges required:**
This tool modifies system network settings and requires elevated privileges to function properly.

🛑 **Platform Compatibility:**
While designed to be cross-platform, some features may have limited functionality on certain platforms due to OS-specific networking implementations.

## 🔧 Installation

### Prerequisites

- Rust 1.70+ (for building from source)
- Platform-specific network tools:
  - **Linux**: `ip`, `iptables`, `tc`
  - **Windows**: `netsh`, `powershell`
  - **macOS**: `networksetup`, `scutil`, `pfctl`

### From Source

1. Clone the repository:
   ```bash
   git clone https://github.com/RimakiTaema/netmgr.git
   cd netmgr
   ```

2. Build for your platform:
   ```bash
   cargo build --release
   ```

3. Install (Linux/macOS):
   ```bash
   sudo cp target/release/netmgr /usr/local/bin/
   ```

### Cross-Compilation

To build for multiple platforms:

```bash
make build-all
```

This will create binaries for various platforms in the `target` directory.

## 📖 Usage

```bash
netmgr <context> <command> [parameters...]
```

### Global Options

- `-v, --verbose`: Enable verbose output
- `-n, --dry-run`: Show what would be done without executing
- `-f, --force`: Force operations without confirmation
- `-h, --help`: Show help

### Contexts

#### Interface Management
```bash
# Show all interfaces
netmgr interface show

# Show specific interface
netmgr interface show eth0

# Set IP address
netmgr interface set eth0 ip 192.168.1.100 24

# Bring interface up/down
netmgr interface set eth0 up
netmgr interface set eth0 down

# Set MTU
netmgr interface set eth0 mtu 1500
```

#### Route Management
```bash
# Show routing table
netmgr route show

# Add route
netmgr route add 10.0.0.0/8 --via 192.168.1.1

# Delete route
netmgr route delete 10.0.0.0/8
```

#### Firewall Management
```bash
# Show firewall rules
netmgr firewall show

# Allow port
netmgr firewall rule allow 22 tcp

# Block port
netmgr firewall rule deny 25 tcp

# Flush rules
netmgr firewall rule flush
```

#### Port Forwarding
```bash
# Show forwards
netmgr forward show

# Add port forward
netmgr forward add web 80 192.168.1.100 8080 --protocol tcp

# Remove forward
netmgr forward remove web
```

#### DNS Configuration
```bash
# Show DNS settings
netmgr dns show

# Set DNS servers
netmgr dns set 8.8.8.8 1.1.1.1
```

#### Bandwidth Management
```bash
# Show bandwidth config
netmgr bandwidth show eth0

# Limit bandwidth
netmgr bandwidth limit eth0 100mbit
```

#### Network Diagnostics
```bash
# Test connectivity
netmgr diag connectivity google.com --count 5

# Test ports
netmgr diag ports 192.168.1.1 22,80,443

# Monitor bandwidth
netmgr diag bandwidth eth0 --duration 30
```

## 🏗️ Architecture

NetMgr is built with a modular architecture:

- **CLI Layer** - Command parsing and user interface
- **Common Layer** - Shared utilities, configuration, and platform detection
- **Module Layer** - Feature-specific implementations
- **Platform Layer** - OS-specific command execution

### Key Features

- **Memory Safety** - Written in Rust for guaranteed memory safety
- **Cross-Platform** - Single codebase supporting Linux, Windows, and macOS
- **Async Operations** - Built on Tokio for efficient I/O operations
- **Configuration Management** - JSON-based configuration with automatic backups
- **Comprehensive Logging** - Detailed logging with multiple levels
- **Dry-Run Mode** - Test commands without making changes

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Inspired by Windows `netsh` command
- Built with the amazing Rust ecosystem
- Thanks to all contributors and testers
