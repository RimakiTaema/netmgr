use anyhow::Result;
use log::info;
use std::collections::HashMap;

use crate::cli::InterfaceCommands;
use crate::common::*;

pub async fn handle_command(command: InterfaceCommands, options: &GlobalOptions) -> Result<()> {
    match command {
        InterfaceCommands::Show { interface } => {
            if let Some(iface) = interface {
                show_interface(&iface, options).await
            } else {
                show_all_interfaces(options).await
            }
        }
        InterfaceCommands::Set { interface, property, values } => {
            set_interface(&interface, &property, &values, options).await
        }
    }
}

async fn show_all_interfaces(options: &GlobalOptions) -> Result<()> {
    log_info("All network interfaces:");
    println!();
    
    println!("{:<15} {:<10} {:<15} {:<20} {:<10}", 
             "INTERFACE", "STATE", "IP ADDRESS", "MAC ADDRESS", "MTU");
    println!("{:<15} {:<10} {:<15} {:<20} {:<10}", 
             "---------", "-----", "----------", "-----------", "---");
    
    let interfaces = get_interfaces().await?;
    
    for (name, iface) in interfaces {
        // Skip loopback
        if name == "lo" || name.starts_with("Loopback") {
            continue;
        }
        
        println!("{:<15} {:<10} {:<15} {:<20} {:<10}", 
                 name, 
                 iface.state, 
                 iface.ip_address.unwrap_or_else(|| "N/A".to_string()),
                 iface.mac_address.unwrap_or_else(|| "N/A".to_string()),
                 iface.mtu.unwrap_or(0));
    }
    
    Ok(())
}

async fn show_interface(interface_name: &str, options: &GlobalOptions) -> Result<()> {
    log_info(&format!("Interface details for: {}", interface_name));
    println!();
    
    let interfaces = get_interfaces().await?;
    let iface = interfaces.get(interface_name)
        .ok_or_else(|| NetMgrError::NetworkError(format!("Interface {} not found", interface_name)))?;
    
    println!("{}=== Interface Information ==={}", COLORS.cyan, COLORS.reset);
    println!("Name: {}", interface_name);
    println!("State: {}", iface.state);
    if let Some(ip) = &iface.ip_address {
        println!("IP Address: {}", ip);
    }
    if let Some(mac) = &iface.mac_address {
        println!("MAC Address: {}", mac);
    }
    if let Some(mtu) = iface.mtu {
        println!("MTU: {}", mtu);
    }
    
    // Platform-specific details
    show_platform_interface_details(interface_name, options).await?;
    
    Ok(())
}

async fn set_interface(interface_name: &str, property: &str, values: &[String], options: &GlobalOptions) -> Result<()> {
    match property {
        "ip" | "address" => {
            if values.is_empty() {
                return Err(NetMgrError::ConfigError("IP address required".to_string()).into());
            }
            let ip_address = &values[0];
            let prefix = values.get(1).map(|s| s.as_str()).unwrap_or("24");
            set_interface_ip(interface_name, ip_address, prefix, options).await
        }
        "up" => set_interface_state(interface_name, true, options).await,
        "down" => set_interface_state(interface_name, false, options).await,
        "mtu" => {
            if values.is_empty() {
                return Err(NetMgrError::ConfigError("MTU value required".to_string()).into());
            }
            set_interface_mtu(interface_name, &values[0], options).await
        }
        "mac" => {
            if values.is_empty() {
                return Err(NetMgrError::ConfigError("MAC address required".to_string()).into());
            }
            set_interface_mac(interface_name, &values[0], options).await
        }
        _ => Err(NetMgrError::ConfigError(format!("Unknown property: {}", property)).into()),
    }
}

#[derive(Debug)]
struct InterfaceInfo {
    state: String,
    ip_address: Option<String>,
    mac_address: Option<String>,
    mtu: Option<u32>,
}

async fn get_interfaces() -> Result<HashMap<String, InterfaceInfo>> {
    #[cfg(target_os = "linux")]
    return get_linux_interfaces().await;
    
    #[cfg(target_os = "windows")]
    return get_windows_interfaces().await;
    
    #[cfg(target_os = "macos")]
    return get_macos_interfaces().await;
    
    #[cfg(not(any(target_os = "linux", target_os = "windows", target_os = "macos")))]
    Err(NetMgrError::UnsupportedPlatform("Interface management not supported".to_string()).into())
}

#[cfg(target_os = "linux")]
async fn get_linux_interfaces() -> Result<HashMap<String, InterfaceInfo>> {
    use std::process::Command;
    
    let output = Command::new("ip")
        .args(&["link", "show"])
        .output()?;
    
    let output_str = String::from_utf8_lossy(&output.stdout);
    let mut interfaces = HashMap::new();
    
    for line in output_str.lines() {
        if line.contains(": ") && !line.starts_with(' ') {
            let parts: Vec<&str> = line.split(": ").collect();
            if parts.len() >= 2 {
                let name = parts[1].split('@').next().unwrap_or(parts[1]).to_string();
                let state = if line.contains("state UP") { "UP" } else { "DOWN" }.to_string();
                
                // Get IP address
                let ip_output = Command::new("ip")
                    .args(&["addr", "show", &name])
                    .output()?;
                let ip_str = String::from_utf8_lossy(&ip_output.stdout);
                let ip_address = extract_ip_from_output(&ip_str);
                
                // Get MAC address
                let mac_address = extract_mac_from_line(line);
                
                // Get MTU
                let mtu = extract_mtu_from_line(line);
                
                interfaces.insert(name, InterfaceInfo {
                    state,
                    ip_address,
                    mac_address,
                    mtu,
                });
            }
        }
    }
    
    Ok(interfaces)
}

#[cfg(target_os = "windows")]
async fn get_windows_interfaces() -> Result<HashMap<String, InterfaceInfo>> {
    use std::process::Command;
    
    let output = Command::new("powershell")
        .args(&["-Command", "Get-NetAdapter | Select-Object Name,Status,LinkSpeed,MacAddress | ConvertTo-Json"])
        .output()?;
    
    let output_str = String::from_utf8_lossy(&output.stdout);
    let mut interfaces = HashMap::new();
    
    // Parse JSON output from PowerShell
    if let Ok(json_value) = serde_json::from_str::<serde_json::Value>(&output_str) {
        if let Some(array) = json_value.as_array() {
            for item in array {
                if let Some(name) = item.get("Name").and_then(|v| v.as_str()) {
                    let state = item.get("Status")
                        .and_then(|v| v.as_str())
                        .unwrap_or("Unknown")
                        .to_string();
                    let mac_address = item.get("MacAddress")
                        .and_then(|v| v.as_str())
                        .map(|s| s.to_string());
                    
                    interfaces.insert(name.to_string(), InterfaceInfo {
                        state,
                        ip_address: None, // Will be filled separately
                        mac_address,
                        mtu: None,
                    });
                }
            }
        }
    }
    
    Ok(interfaces)
}

#[cfg(target_os = "macos")]
async fn get_macos_interfaces() -> Result<HashMap<String, InterfaceInfo>> {
    use std::process::Command;
    
    let output = Command::new("ifconfig")
        .output()?;
    
    let output_str = String::from_utf8_lossy(&output.stdout);
    let mut interfaces = HashMap::new();
    let mut current_interface = None;
    
    for line in output_str.lines() {
        if !line.starts_with('\t') && !line.starts_with(' ') && line.contains(':') {
            // New interface
            let name = line.split(':').next().unwrap_or("").to_string();
            let state = if line.contains("UP") { "UP" } else { "DOWN" }.to_string();
            current_interface = Some(name.clone());
            
            interfaces.insert(name, InterfaceInfo {
                state,
                ip_address: None,
                mac_address: None,
                mtu: None,
            });
        } else if let Some(ref iface_name) = current_interface {
            if let Some(iface) = interfaces.get_mut(iface_name) {
                if line.contains("inet ") && !line.contains("127.0.0.1") {
                    if let Some(ip) = extract_ip_from_output(line) {
                        iface.ip_address = Some(ip);
                    }
                }
                if line.contains("ether ") {
                    if let Some(mac) = extract_mac_from_line(line) {
                        iface.mac_address = Some(mac);
                    }
                }
                if line.contains("mtu ") {
                    if let Some(mtu) = extract_mtu_from_line(line) {
                        iface.mtu = Some(mtu);
                    }
                }
            }
        }
    }
    
    Ok(interfaces)
}

fn extract_ip_from_output(output: &str) -> Option<String> {
    for line in output.lines() {
        if line.contains("inet ") && !line.contains("127.0.0.1") {
            let parts: Vec<&str> = line.split_whitespace().collect();
            for (i, part) in parts.iter().enumerate() {
                if *part == "inet" && i + 1 < parts.len() {
                    return Some(parts[i + 1].split('/').next().unwrap_or(parts[i + 1]).to_string());
                }
            }
        }
    }
    None
}

fn extract_mac_from_line(line: &str) -> Option<String> {
    // Look for MAC address pattern (xx:xx:xx:xx:xx:xx)
    let words: Vec<&str> = line.split_whitespace().collect();
    for word in words {
        if word.matches(':').count() == 5 && word.len() == 17 {
            return Some(word.to_string());
        }
    }
    None
}

fn extract_mtu_from_line(line: &str) -> Option<u32> {
    let words: Vec<&str> = line.split_whitespace().collect();
    for (i, word) in words.iter().enumerate() {
        if *word == "mtu" && i + 1 < words.len() {
            return words[i + 1].parse().ok();
        }
    }
    None
}

async fn show_platform_interface_details(interface_name: &str, options: &GlobalOptions) -> Result<()> {
    #[cfg(target_os = "linux")]
    {
        println!("\n{}=== Statistics ==={}", COLORS.cyan, COLORS.reset);
        if let Ok(output) = execute_command("ip", &["-s", "link", "show", interface_name], options).await {
            println!("{}", output);
        }
        
        println!("\n{}=== Routes ==={}", COLORS.cyan, COLORS.reset);
        if let Ok(output) = execute_command("ip", &["route", "show", "dev", interface_name], options).await {
            println!("{}", output);
        }
    }
    
    #[cfg(target_os = "windows")]
    {
        println!("\n{}=== Interface Details ==={}", COLORS.cyan, COLORS.reset);
        if let Ok(output) = execute_command("netsh", &["interface", "ip", "show", "addresses", interface_name], options).await {
            println!("{}", output);
        }
    }
    
    #[cfg(target_os = "macos")]
    {
        println!("\n{}=== Interface Details ==={}", COLORS.cyan, COLORS.reset);
        if let Ok(output) = execute_command("ifconfig", &[interface_name], options).await {
            println!("{}", output);
        }
    }
    
    Ok(())
}

async fn set_interface_ip(interface_name: &str, ip_address: &str, prefix: &str, options: &GlobalOptions) -> Result<()> {
    log_info(&format!("Setting IP address {}/{} on {}", ip_address, prefix, interface_name));
    
    #[cfg(target_os = "linux")]
    {
        execute_command("ip", &["addr", "add", &format!("{}/{}", ip_address, prefix), "dev", interface_name], options).await?;
    }
    
    #[cfg(target_os = "windows")]
    {
        execute_command("netsh", &["interface", "ip", "set", "address", interface_name, "static", ip_address, prefix], options).await?;
    }
    
    #[cfg(target_os = "macos")]
    {
        execute_command("ifconfig", &[interface_name, "inet", &format!("{}/{}", ip_address, prefix)], options).await?;
    }
    
    Ok(())
}

async fn set_interface_state(interface_name: &str, up: bool, options: &GlobalOptions) -> Result<()> {
    let state_str = if up { "up" } else { "down" };
    log_info(&format!("Bringing interface {} {}", interface_name, state_str));
    
    #[cfg(target_os = "linux")]
    {
        execute_command("ip", &["link", "set", interface_name, state_str], options).await?;
    }
    
    #[cfg(target_os = "windows")]
    {
        let action = if up { "enable" } else { "disable" };
        execute_command("netsh", &["interface", "set", "interface", interface_name, action], options).await?;
    }
    
    #[cfg(target_os = "macos")]
    {
        execute_command("ifconfig", &[interface_name, state_str], options).await?;
    }
    
    Ok(())
}

async fn set_interface_mtu(interface_name: &str, mtu: &str, options: &GlobalOptions) -> Result<()> {
    log_info(&format!("Setting MTU to {} on {}", mtu, interface_name));
    
    #[cfg(target_os = "linux")]
    {
        execute_command("ip", &["link", "set", interface_name, "mtu", mtu], options).await?;
    }
    
    #[cfg(target_os = "windows")]
    {
        execute_command("netsh", &["interface", "ipv4", "set", "subinterface", interface_name, &format!("mtu={}", mtu)], options).await?;
    }
    
    #[cfg(target_os = "macos")]
    {
        execute_command("ifconfig", &[interface_name, "mtu", mtu], options).await?;
    }
    
    Ok(())
}

async fn set_interface_mac(interface_name: &str, mac: &str, options: &GlobalOptions) -> Result<()> {
    log_info(&format!("Setting MAC address to {} on {}", mac, interface_name));
    
    #[cfg(target_os = "linux")]
    {
        execute_command("ip", &["link", "set", interface_name, "down"], options).await?;
        execute_command("ip", &["link", "set", interface_name, "address", mac], options).await?;
        execute_command("ip", &["link", "set", interface_name, "up"], options).await?;
    }
    
    #[cfg(target_os = "windows")]
    {
        let mac_no_colons = mac.replace(":", "");
        execute_command("powershell", &["-Command", &format!("Set-NetAdapter -Name '{}' -MacAddress '{}'", interface_name, mac_no_colons)], options).await?;
    }
    
    #[cfg(target_os = "macos")]
    {
        execute_command("ifconfig", &[interface_name, "ether", mac], options).await?;
    }
    
    Ok(())
}
