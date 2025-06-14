use anyhow::Result;
use clap::{Parser, Subcommand};
use log::{error, info};
use std::process;

mod cli;
mod common;
mod modules;

use cli::*;
use common::*;

#[derive(Parser)]
#[command(name = "netmgr")]
#[command(about = "Cross-platform network management tool")]
#[command(version = "1.0.0")]
struct Cli {
    #[command(subcommand)]
    command: Commands,
    
    /// Enable verbose output
    #[arg(short, long, global = true)]
    verbose: bool,
    
    /// Show what would be done without executing
    #[arg(short = 'n', long, global = true)]
    dry_run: bool,
    
    /// Force operations without confirmation
    #[arg(short, long, global = true)]
    force: bool,
}

#[derive(Subcommand)]
enum Commands {
    /// Network interface management
    #[command(alias = "int")]
    Interface {
        #[command(subcommand)]
        command: InterfaceCommands,
    },
    /// Routing table management
    #[command(alias = "rt")]
    Route {
        #[command(subcommand)]
        command: RouteCommands,
    },
    /// Firewall rules management
    #[command(alias = "fw")]
    Firewall {
        #[command(subcommand)]
        command: FirewallCommands,
    },
    /// Port forwarding management
    #[command(alias = "fwd")]
    Forward {
        #[command(subcommand)]
        command: ForwardCommands,
    },
    /// DNS configuration
    Dns {
        #[command(subcommand)]
        command: DnsCommands,
    },
    /// Traffic shaping and QoS
    #[command(alias = "bw")]
    Bandwidth {
        #[command(subcommand)]
        command: BandwidthCommands,
    },
    /// Tunnel interfaces
    #[command(alias = "tun")]
    Tunnel {
        #[command(subcommand)]
        command: TunnelCommands,
    },
    /// Network diagnostics
    #[command(alias = "diagnostic")]
    Diag {
        #[command(subcommand)]
        command: DiagCommands,
    },
}

#[tokio::main]
async fn main() -> Result<()> {
    let cli = Cli::parse();
    
    // Initialize logging
    let log_level = if cli.verbose { "debug" } else { "info" };
    env_logger::Builder::from_env(env_logger::Env::default().default_filter_or(log_level))
        .init();
    
    // Initialize global options
    let options = GlobalOptions {
        verbose: cli.verbose,
        dry_run: cli.dry_run,
        force: cli.force,
    };
    
    // Initialize system
    if let Err(e) = common::init_system().await {
        error!("Failed to initialize system: {}", e);
        process::exit(1);
    }
    
    // Check for root privileges on Unix systems
    #[cfg(unix)]
    if !options.dry_run && !common::is_root() {
        error!("This tool requires administrator privileges");
        process::exit(1);
    }
    
    // Check dependencies
    if let Err(e) = common::check_dependencies().await {
        error!("Dependency check failed: {}", e);
        process::exit(1);
    }
    
    // Handle commands
    let result = match cli.command {
        Commands::Interface { command } => {
            modules::interface::handle_command(command, &options).await
        }
        Commands::Route { command } => {
            modules::route::handle_command(command, &options).await
        }
        Commands::Firewall { command } => {
            modules::firewall::handle_command(command, &options).await
        }
        Commands::Forward { command } => {
            modules::forward::handle_command(command, &options).await
        }
        Commands::Dns { command } => {
            modules::dns::handle_command(command, &options).await
        }
        Commands::Bandwidth { command } => {
            modules::bandwidth::handle_command(command, &options).await
        }
        Commands::Tunnel { command } => {
            modules::tunnel::handle_command(command, &options).await
        }
        Commands::Diag { command } => {
            modules::diagnostic::handle_command(command, &options).await
        }
    };
    
    if let Err(e) = result {
        error!("Command failed: {}", e);
        process::exit(1);
    }
    
    Ok(())
}
