use anyhow::{Context, Result};
use log::{debug, error, info, warn};
use serde::{Deserialize, Serialize};
use std::path::PathBuf;
use std::process::Command;
use tokio::fs;

pub mod colors;
pub mod config;
pub mod platform;

pub use colors::*;
pub use config::*;
pub use platform::*;

#[derive(Debug, Clone)]
pub struct GlobalOptions {
    pub verbose: bool,
    pub dry_run: bool,
    pub force: bool,
}

#[derive(Debug, thiserror::Error)]
pub enum NetMgrError {
    #[error("Command execution failed: {0}")]
    CommandFailed(String),
    #[error("Platform not supported: {0}")]
    UnsupportedPlatform(String),
    #[error("Configuration error: {0}")]
    ConfigError(String),
    #[error("Network operation failed: {0}")]
    NetworkError(String),
}

pub async fn init_system() -> Result<()> {
    info!("Initializing netmgr system...");
    
    // Create config directories
    let config_dir = get_config_dir()?;
    let state_dir = get_state_dir()?;
    let log_dir = get_log_dir()?;
    
    fs::create_dir_all(&config_dir).await
        .context("Failed to create config directory")?;
    fs::create_dir_all(&state_dir).await
        .context("Failed to create state directory")?;
    fs::create_dir_all(&log_dir).await
        .context("Failed to create log directory")?;
    
    // Initialize default config files
    init_default_configs(&config_dir).await?;
    
    info!("System initialized successfully");
    Ok(())
}

async fn init_default_configs(config_dir: &PathBuf) -> Result<()> {
    let default_configs = [
        "interfaces.json",
        "forwarding.json", 
        "firewall.json",
        "routing.json",
        "dns.json",
        "tunnels.json",
    ];
    
    for config_file in &default_configs {
        let config_path = config_dir.join(config_file);
        if !config_path.exists() {
            fs::write(&config_path, "{}")
                .await
                .context(format!("Failed to create {}", config_file))?;
        }
    }
    
    Ok(())
}

pub async fn execute_command(
    command: &str,
    args: &[&str],
    options: &GlobalOptions,
) -> Result<String> {
    let cmd_str = format!("{} {}", command, args.join(" "));
    debug!("Executing: {}", cmd_str);
    
    if options.dry_run {
        info!("[DRY-RUN] Would execute: {}", cmd_str);
        return Ok(String::new());
    }
    
    let output = Command::new(command)
        .args(args)
        .output()
        .context(format!("Failed to execute command: {}", cmd_str))?;
    
    if !output.status.success() {
        let stderr = String::from_utf8_lossy(&output.stderr);
        error!("Command failed: {}", cmd_str);
        error!("Error: {}", stderr);
        return Err(NetMgrError::CommandFailed(stderr.to_string()).into());
    }
    
    let stdout = String::from_utf8_lossy(&output.stdout).to_string();
    debug!("Command successful");
    Ok(stdout)
}

pub fn log_info(message: &str) {
    println!("{}{}{}", COLORS.green, message, COLORS.reset);
}

pub fn log_error(message: &str) {
    eprintln!("{}{}{}", COLORS.red, message, COLORS.reset);
}

pub fn log_warn(message: &str) {
    println!("{}{}{}", COLORS.yellow, message, COLORS.reset);
}

pub fn log_debug(message: &str) {
    println!("{}{}{}", COLORS.gray, message, COLORS.reset);
}

#[cfg(unix)]
pub fn is_root() -> bool {
    unsafe { libc::geteuid() == 0 }
}

#[cfg(windows)]
pub fn is_root() -> bool {
    // On Windows, check if we can write to a protected directory
    use std::fs::File;
    use std::path::Path;
    
    let test_path = Path::new("C:\\Windows\\temp_netmgr_test");
    match File::create(test_path) {
        Ok(_) => {
            let _ = std::fs::remove_file(test_path);
            true
        }
        Err(_) => false,
    }
}

pub async fn check_dependencies() -> Result<()> {
    let required_tools = get_required_tools();
    
    for tool in required_tools {
        if which::which(&tool).is_err() {
            return Err(NetMgrError::ConfigError(
                format!("Required tool not found: {}", tool)
            ).into());
        }
    }
    
    Ok(())
}

fn get_required_tools() -> Vec<String> {
    #[cfg(target_os = "linux")]
    return vec!["ip".to_string(), "iptables".to_string(), "tc".to_string()];
    
    #[cfg(target_os = "windows")]
    return vec!["netsh".to_string(), "powershell".to_string()];
    
    #[cfg(target_os = "macos")]
    return vec!["networksetup".to_string(), "scutil".to_string(), "pfctl".to_string()];
    
    #[cfg(not(any(target_os = "linux", target_os = "windows", target_os = "macos")))]
    return vec![];
}
