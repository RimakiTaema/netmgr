use clap::Subcommand;

#[derive(Subcommand)]
pub enum InterfaceCommands {
    /// Show all interfaces or details of specific interface
    Show {
        /// Interface name to show details for
        interface: Option<String>,
    },
    /// Configure interface properties
    Set {
        /// Interface name
        interface: String,
        /// Property to set (ip, up, down, mtu, mac)
        property: String,
        /// Property value(s)
        values: Vec<String>,
    },
}

#[derive(Subcommand)]
pub enum RouteCommands {
    /// Show routing table
    Show {
        /// Route type (table, cache, all)
        route_type: Option<String>,
    },
    /// Add a route
    Add {
        /// Destination network
        destination: String,
        /// Gateway (via)
        #[arg(long)]
        via: Option<String>,
        /// Interface (dev)
        #[arg(long)]
        dev: Option<String>,
        /// Routing table
        #[arg(long)]
        table: Option<String>,
    },
    /// Delete a route
    Delete {
        /// Destination network
        destination: String,
        /// Routing table
        #[arg(long)]
        table: Option<String>,
    },
}

#[derive(Subcommand)]
pub enum FirewallCommands {
    /// Show firewall rules
    Show {
        /// Table name (filter, nat, mangle, raw)
        table: Option<String>,
    },
    /// Manage firewall rules
    Rule {
        /// Action (allow, deny, flush, save, restore)
        action: String,
        /// Rule parameters
        params: Vec<String>,
    },
}

#[derive(Subcommand)]
pub enum ForwardCommands {
    /// Show active port forwards
    Show,
    /// Add a port forward
    Add {
        /// Forward name
        name: String,
        /// Source port
        src_port: String,
        /// Destination IP
        dest_ip: String,
        /// Destination port
        dest_port: String,
        /// Protocol (tcp/udp)
        #[arg(short, long, default_value = "tcp")]
        protocol: String,
    },
    /// Remove a port forward
    Remove {
        /// Forward name
        name: String,
    },
}

#[derive(Subcommand)]
pub enum DnsCommands {
    /// Show current DNS configuration
    Show,
    /// Set DNS servers
    Set {
        /// Primary DNS server
        primary: String,
        /// Secondary DNS server
        secondary: Option<String>,
    },
}

#[derive(Subcommand)]
pub enum BandwidthCommands {
    /// Show bandwidth configuration
    Show {
        /// Interface name
        interface: Option<String>,
    },
    /// Set bandwidth limit
    Limit {
        /// Interface name
        interface: String,
        /// Rate limit
        rate: String,
        /// Burst limit
        burst: Option<String>,
    },
}

#[derive(Subcommand)]
pub enum TunnelCommands {
    /// Create a new tunnel
    Create {
        /// Tunnel name
        name: String,
        /// Tunnel type (gre, ipip, sit)
        tunnel_type: String,
        /// Local IP address
        local_ip: String,
        /// Remote IP address
        remote_ip: String,
    },
    /// Delete a tunnel
    Delete {
        /// Tunnel name
        name: String,
    },
}

#[derive(Subcommand)]
pub enum DiagCommands {
    /// Test connectivity to target
    Connectivity {
        /// Target host
        #[arg(default_value = "8.8.8.8")]
        target: String,
        /// Number of pings
        #[arg(short, long, default_value = "3")]
        count: String,
    },
    /// Test if ports are open on target
    Ports {
        /// Target host
        target: String,
        /// Comma-separated list of ports
        #[arg(default_value = "22,80,443,25565")]
        ports: String,
    },
    /// Monitor bandwidth on interface
    Bandwidth {
        /// Interface name
        #[arg(default_value = "eth0")]
        interface: String,
        /// Duration in seconds
        #[arg(short, long, default_value = "10")]
        duration: String,
    },
}
