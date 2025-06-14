use anyhow::Result;
use crate::cli::TunnelCommands;
use crate::common::*;

pub async fn handle_command(command: TunnelCommands, options: &GlobalOptions) -> Result<()> {
    match command {
        TunnelCommands::Create { name, tunnel_type, local_ip, remote_ip } => {
            create_tunnel(&name, &tunnel_type, &local_ip, &remote_ip, options).await
        }
        TunnelCommands::Delete { name } => {
            delete_tunnel(&name, options).await
        }
    }
}

async fn create_tunnel(name: &str, tunnel_type: &str, local_ip: &str, remote_ip: &str, options: &GlobalOptions) -> Result<()> {
    log_info(&format!("Creating {} tunnel: {}", tunnel_type, name));
    
    #[cfg(target_os = "linux")]
    {
        let mode = match tunnel_type {
            "gre" => "gre",
            "ipip" => "ipip",
            "sit" => "sit",
            _ => return Err(NetMgrError::ConfigError(format!("Unsupported tunnel type: {}", tunnel_type)).into()),
        };
        
        execute_command("ip", &["tunnel", "add", name, "mode", mode, 
                               "remote", remote_ip, "local", local_ip], options).await?;
        
        execute_command("ip", &["link", "set", name, "up"], options).await?;
    }
    
    #[cfg(target_os = "windows")]
    {
        match tunnel_type {
            "gre" => {
                execute_command("netsh", &["interface", "ipv4", "add", "interface", 
                                          name, "type=tunnel", &format!("source={}", local_ip), 
                                          &format!("destination={}", remote_ip)], options).await?;
            }
            "ipip" | "sit" => {
                return Err(NetMgrError::UnsupportedPlatform(format!("Tunnel type {} not supported on Windows", tunnel_type)).into());
            }
            _ => {
                return Err(NetMgrError::ConfigError(format!("Unsupported tunnel type: {}", tunnel_type)).into());
            }
        }
    }
    
    #[cfg(target_os = "macos")]
    {
        return Err(NetMgrError::UnsupportedPlatform("Creating tunnels not implemented for macOS".to_string()).into());
    }
    
    Ok(())
}

async fn delete_tunnel(name: &str, options: &GlobalOptions) -> Result<()> {
    log_info(&format!("Deleting tunnel: {}", name));
    
    #[cfg(target_os = "linux")]
    {
        execute_command("ip", &["link", "set", name, "down"], options).await?;
        execute_command("ip", &["tunnel", "del", name], options).await?;
    }
    
    #[cfg(target_os = "windows")]
    {
        execute_command("netsh", &["interface", "ipv4", "delete", "interface", name], options).await?;
    }
    
    #[cfg(target_os = "macos")]
    {
        return Err(NetMgrError::UnsupportedPlatform("Deleting tunnels not implemented for macOS".to_string()).into());
    }
    
    Ok(())
}
