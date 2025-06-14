use anyhow::Result;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use chrono::{DateTime, Utc};

use crate::cli::ForwardCommands;
use crate::common::*;

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct PortForward {
    pub name: String,
    pub src_port: String,
    pub dest_ip: String,
    pub dest_port: String,
    pub protocol: String,
    pub created: DateTime<Utc>,
    pub active: bool,
}

pub type ForwardConfig = HashMap<String, PortForward>;

pub async fn handle_command(command: ForwardCommands, options: &GlobalOptions) -> Result<()> {
    match command {
        ForwardCommands::Show => show_forwards(options).await,
        ForwardCommands::Add { name, src_port, dest_ip, dest_port, protocol } => {
            add_forward(&name, &src_port, &dest_ip, &dest_port, &protocol, options).await
        }
        ForwardCommands::Remove { name } => remove_forward(&name, options).await,
    }
}

async fn show_forwards(options: &GlobalOptions) -> Result<()> {
    log_info("Active port forwards:");
    println!();
    
    println!("{:<15} {:<10} {:<25} {:<10} {:<20}", 
             "NAME", "PROTOCOL", "FORWARD", "STATUS", "CREATED");
    println!("{:<15} {:<10} {:<25} {:<10} {:<20}", 
             "----", "--------", "-------", "------", "-------");
    
    let config = load_forward_config().await?;
    
    for (name, forward) in config {
        let status = if forward.active { "ACTIVE" } else { "INACTIVE" };
        let forward_str = format!("{}->{}:{}", forward.src_port, forward.dest_ip, forward.dest_port);
        let created = forward.created.format("%Y-%m-%d %H:%M:%S").to_string();
        
        println!("{:<15} {:<10} {:<25} {:<10} {:<20}", 
                 name, forward.protocol, forward_str, status, created);
    }
    
    Ok(())
}

async fn add_forward(name: &str, src_port: &str, dest_ip: &str, dest_port: &str, protocol: &str, options: &GlobalOptions) -> Result<()> {
    let mut config = load_forward_config().await?;
    
    if config.contains_key(name) {
        return Err(NetMgrError::ConfigError(format!("Port forward with name '{}' already exists", name)).into());
    }
    
    log_info(&format!("Adding port forward: {} ({}:{} -> {}:{})", name, protocol, src_port, dest_ip, dest_port));
    
    // Enable IP forwarding
    enable_ip_forwarding(options).await?;
    
    // Add platform-specific forwarding rules
    let success = add_platform_forward(name, src_port, dest_ip, dest_port, protocol, options).await?;
    
    if success {
        let forward = PortForward {
            name: name.to_string(),
            src_port: src_port.to_string(),
            dest_ip: dest_ip.to_string(),
            dest_port: dest_port.to_string(),
            protocol: protocol.to_string(),
            created: Utc::now(),
            active: true,
        };
        
        config.insert(name.to_string(), forward);
        save_forward_config(&config).await?;
    }
    
    Ok(())
}

async fn remove_forward(name: &str, options: &GlobalOptions) -> Result<()> {
    let mut config = load_forward_config().await?;
    
    let forward = config.get(name)
        .ok_or_else(|| NetMgrError::ConfigError(format!("Port forward with name '{}' does not exist", name)))?
        .clone();
    
    log_info(&format!("Removing port forward: {}", name));
    
    remove_platform_forward(name, &forward.src_port, &forward.dest_ip, &forward.dest_port, &forward.protocol, options).await?;
    
    config.remove(name);
    save_forward_config(&config).await?;
    
    Ok(())
}

async fn load_forward_config() -> Result<ForwardConfig> {
    match load_config::<ForwardConfig>("forwarding.json").await {
        Ok(config) => Ok(config),
        Err(_) => Ok(HashMap::new()),
    }
}

async fn save_forward_config(config: &ForwardConfig) -> Result<()> {
    save_config("forwarding.json", config).await
}

async fn enable_ip_forwarding(options: &GlobalOptions) -> Result<()> {
    #[cfg(target_os = "linux")]
    {
        execute_command("sysctl", &["-w", "net.ipv4.ip_forward=1"], options).await?;
    }
    
    #[cfg(target_os = "macos")]
    {
        execute_command("sysctl", &["-w", "net.inet.ip.forwarding=1"], options).await?;
    }
    
    // Windows enables forwarding per interface, handled in add_platform_forward
    
    Ok(())
}

async fn add_platform_forward(name: &str, src_port: &str, dest_ip: &str, dest_port: &str, protocol: &str, options: &GlobalOptions) -> Result<bool> {
    #[cfg(target_os = "linux")]
    {
        // Add DNAT rule
        execute_command("iptables", &["-t", "nat", "-A", "PREROUTING", 
                                     "-p", protocol, "--dport", src_port, 
                                     "-j", "DNAT", &format!("--to-destination={}:{}", dest_ip, dest_port),
                                     "-m", "comment", &format!("--comment=NETMGR:{}", name)], options).await?;
        
        // Add FORWARD rule
        execute_command("iptables", &["-A", "FORWARD", 
                                     "-p", protocol, "-d", dest_ip, "--dport", dest_port,
                                     "-j", "ACCEPT",
                                     "-m", "comment", &format!("--comment=NETMGR:{}", name)], options).await?;
        
        // Add MASQUERADE rule
        execute_command("iptables", &["-t", "nat", "-A", "POSTROUTING", 
                                     "-p", protocol, "-d", dest_ip, "--dport", dest_port,
                                     "-j", "MASQUERADE",
                                     "-m", "comment", &format!("--comment=NETMGR:{}", name)], options).await?;
        
        return Ok(true);
    }
    
    #[cfg(target_os = "windows")]
    {
        // Windows uses netsh portproxy
        execute_command("netsh", &["interface", "portproxy", "add", "v4tov4", 
                                  &format!("listenport={}", src_port), "listenaddress=0.0.0.0",
                                  &format!("connectport={}", dest_port), &format!("connectaddress={}", dest_ip),
                                  &format!("protocol={}", protocol)], options).await?;
        
        // Add firewall rule to allow incoming traffic
        let rule_name = format!("NetMgr-Forward-{}", name);
        execute_command("netsh", &["advfirewall", "firewall", "add", "rule",
                                  &format!("name={}", rule_name), "dir=in", "action=allow",
                                  &format!("protocol={}", protocol), &format!("localport={}", src_port)], options).await?;
        
        return Ok(true);
    }
    
    #[cfg(target_os = "macos")]
    {
        // macOS uses pfctl for port forwarding
        let rule = format!("rdr pass on lo0 proto {} from any to any port {} -> {} port {}", 
                          protocol, src_port, dest_ip, dest_port);
        
        execute_command("sh", &["-c", &format!("echo '{}' | pfctl -a com.netmgr/{} -f -", rule, name)], options).await?;
        
        // Enable pf if not already enabled
        let _ = execute_command("pfctl", &["-e"], options).await; // Ignore errors as it might already be enabled
        
        return Ok(true);
    }
    
    #[cfg(not(any(target_os = "linux", target_os = "windows", target_os = "macos")))]
    {
        return Err(NetMgrError::UnsupportedPlatform("Port forwarding not implemented for this platform".to_string()).into());
    }
}

async fn remove_platform_forward(name: &str, src_port: &str, dest_ip: &str, dest_port: &str, protocol: &str, options: &GlobalOptions) -> Result<()> {
    #[cfg(target_os = "linux")]
    {
        // Use grep to find and remove rules with the specific comment
        execute_command("sh", &["-c", &format!("iptables-save | grep -v 'NETMGR:{}' | iptables-restore", name)], options).await?;
    }
    
    #[cfg(target_os = "windows")]
    {
        // Remove portproxy rule
        execute_command("netsh", &["interface", "portproxy", "delete", "v4tov4",
                                  &format!("listenport={}", src_port), "listenaddress=0.0.0.0",
                                  &format!("protocol={}", protocol)], options).await?;
        
        // Remove firewall rule
        let rule_name = format!("NetMgr-Forward-{}", name);
        execute_command("netsh", &["advfirewall", "firewall", "delete", "rule",
                                  &format!("name={}", rule_name)], options).await?;
    }
    
    #[cfg(target_os = "macos")]
    {
        // Remove the anchor
        execute_command("pfctl", &["-a", &format!("com.netmgr/{}", name), "-F", "all"], options).await?;
    }
    
    Ok(())
}
