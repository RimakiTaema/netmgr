use anyhow::Result;
use crate::cli::FirewallCommands;
use crate::common::*;

pub async fn handle_command(command: FirewallCommands, options: &GlobalOptions) -> Result<()> {
    match command {
        FirewallCommands::Show { table } => {
            let tbl = table.as_deref().unwrap_or("filter");
            show_firewall(tbl, options).await
        }
        FirewallCommands::Rule { action, params } => {
            handle_firewall_rule(&action, &params, options).await
        }
    }
}

async fn show_firewall(table: &str, options: &GlobalOptions) -> Result<()> {
    log_info(&format!("Firewall rules (table: {}):", table));
    println!();
    
    #[cfg(target_os = "linux")]
    {
        let args = vec!["-t", table, "-L", "-n", "-v", "--line-numbers"];
        let output = execute_command("iptables", &args, options).await?;
        println!("{}", output);
    }
    
    #[cfg(target_os = "windows")]
    {
        let output = execute_command("netsh", &["advfirewall", "firewall", "show", "rule", "name=all"], options).await?;
        println!("{}", output);
    }
    
    #[cfg(target_os = "macos")]
    {
        let output = execute_command("pfctl", &["-s", "rules"], options).await?;
        println!("{}", output);
    }
    
    Ok(())
}

async fn handle_firewall_rule(action: &str, params: &[String], options: &GlobalOptions) -> Result<()> {
    match action {
        "allow" => {
            if params.is_empty() {
                return Err(NetMgrError::ConfigError("Usage: firewall rule allow <port> [protocol] [interface]".to_string()).into());
            }
            let port = &params[0];
            let protocol = params.get(1).map(|s| s.as_str()).unwrap_or("tcp");
            let interface = params.get(2).map(|s| s.as_str());
            
            allow_port(port, protocol, interface, options).await
        }
        "deny" | "block" => {
            if params.is_empty() {
                return Err(NetMgrError::ConfigError("Usage: firewall rule deny <port> [protocol] [interface]".to_string()).into());
            }
            let port = &params[0];
            let protocol = params.get(1).map(|s| s.as_str()).unwrap_or("tcp");
            let interface = params.get(2).map(|s| s.as_str());
            
            deny_port(port, protocol, interface, options).await
        }
        "flush" => {
            let table = params.get(0).map(|s| s.as_str()).unwrap_or("filter");
            flush_firewall(table, options).await
        }
        "save" => {
            let file = params.get(0).map(|s| s.as_str()).unwrap_or(get_default_save_file());
            save_firewall(file, options).await
        }
        "restore" => {
            let file = params.get(0).map(|s| s.as_str()).unwrap_or(get_default_save_file());
            restore_firewall(file, options).await
        }
        _ => Err(NetMgrError::ConfigError(format!("Unknown firewall rule action: {}", action)).into()),
    }
}

async fn allow_port(port: &str, protocol: &str, interface: Option<&str>, options: &GlobalOptions) -> Result<()> {
    log_info(&format!("Adding allow rule for port {}/{}", port, protocol));
    
    #[cfg(target_os = "linux")]
    {
        let mut args = vec!["-A", "INPUT", "-p", protocol, "--dport", port, "-j", "ACCEPT"];
        if let Some(iface) = interface {
            args = vec!["-A", "INPUT", "-i", iface, "-p", protocol, "--dport", port, "-j", "ACCEPT"];
        }
        
        execute_command("iptables", &args, options).await?;
    }
    
    #[cfg(target_os = "windows")]
    {
        let name = format!("NetMgr-Allow-{}-{}", protocol, port);
        let mut args = vec!["advfirewall", "firewall", "add", "rule", 
                           &format!("name={}", name),
                           &format!("protocol={}", protocol),
                           &format!("localport={}", port),
                           "dir=in", "action=allow"];
        
        if let Some(iface) = interface {
            args.push(&format!("interface={}", iface));
        }
        
        execute_command("netsh", &args, options).await?;
    }
    
    #[cfg(target_os = "macos")]
    {
        let rule = if let Some(iface) = interface {
            format!("pass in on {} proto {} from any to any port {}", iface, protocol, port)
        } else {
            format!("pass in proto {} from any to any port {}", protocol, port)
        };
        
        // This is simplified - real implementation would need proper pfctl handling
        execute_command("sh", &["-c", &format!("echo '{}' | pfctl -a com.netmgr/rules -f -", rule)], options).await?;
    }
    
    Ok(())
}

async fn deny_port(port: &str, protocol: &str, interface: Option<&str>, options: &GlobalOptions) -> Result<()> {
    log_info(&format!("Adding deny rule for port {}/{}", port, protocol));
    
    #[cfg(target_os = "linux")]
    {
        let mut args = vec!["-A", "INPUT", "-p", protocol, "--dport", port, "-j", "DROP"];
        if let Some(iface) = interface {
            args = vec!["-A", "INPUT", "-i", iface, "-p", protocol, "--dport", port, "-j", "DROP"];
        }
        
        execute_command("iptables", &args, options).await?;
    }
    
    #[cfg(target_os = "windows")]
    {
        let name = format!("NetMgr-Block-{}-{}", protocol, port);
        let mut args = vec!["advfirewall", "firewall", "add", "rule", 
                           &format!("name={}", name),
                           &format!("protocol={}", protocol),
                           &format!("localport={}", port),
                           "dir=in", "action=block"];
        
        if let Some(iface) = interface {
            args.push(&format!("interface={}", iface));
        }
        
        execute_command("netsh", &args, options).await?;
    }
    
    #[cfg(target_os = "macos")]
    {
        let rule = if let Some(iface) = interface {
            format!("block in on {} proto {} from any to any port {}", iface, protocol, port)
        } else {
            format!("block in proto {} from any to any port {}", protocol, port)
        };
        
        execute_command("sh", &["-c", &format!("echo '{}' | pfctl -a com.netmgr/rules -f -", rule)], options).await?;
    }
    
    Ok(())
}

async fn flush_firewall(table: &str, options: &GlobalOptions) -> Result<()> {
    log_info(&format!("Flushing {} table", table));
    
    #[cfg(target_os = "linux")]
    {
        execute_command("iptables", &["-t", table, "-F"], options).await?;
    }
    
    #[cfg(target_os = "windows")]
    {
        execute_command("netsh", &["advfirewall", "reset"], options).await?;
    }
    
    #[cfg(target_os = "macos")]
    {
        execute_command("pfctl", &["-F", "rules"], options).await?;
    }
    
    Ok(())
}

async fn save_firewall(file: &str, options: &GlobalOptions) -> Result<()> {
    log_info(&format!("Saving firewall rules to {}", file));
    
    #[cfg(target_os = "linux")]
    {
        execute_command("sh", &["-c", &format!("iptables-save > {}", file)], options).await?;
    }
    
    #[cfg(target_os = "windows")]
    {
        execute_command("netsh", &["advfirewall", "export", file], options).await?;
    }
    
    #[cfg(target_os = "macos")]
    {
        execute_command("sh", &["-c", &format!("pfctl -sr > {}", file)], options).await?;
    }
    
    Ok(())
}

async fn restore_firewall(file: &str, options: &GlobalOptions) -> Result<()> {
    log_info(&format!("Restoring firewall rules from {}", file));
    
    #[cfg(target_os = "linux")]
    {
        execute_command("sh", &["-c", &format!("iptables-restore < {}", file)], options).await?;
    }
    
    #[cfg(target_os = "windows")]
    {
        execute_command("netsh", &["advfirewall", "import", file], options).await?;
    }
    
    #[cfg(target_os = "macos")]
    {
        execute_command("sh", &["-c", &format!("pfctl -f {}", file)], options).await?;
    }
    
    Ok(())
}

fn get_default_save_file() -> &'static str {
    #[cfg(target_os = "windows")]
    return "C:\\Windows\\System32\\firewall_rules.wfw";
    
    #[cfg(not(target_os = "windows"))]
    return "/etc/iptables/rules.v4";
}
