use anyhow::Result;
use crate::cli::RouteCommands;
use crate::common::*;

pub async fn handle_command(command: RouteCommands, options: &GlobalOptions) -> Result<()> {
    match command {
        RouteCommands::Show { route_type } => {
            let rt = route_type.as_deref().unwrap_or("table");
            show_routes(rt, options).await
        }
        RouteCommands::Add { destination, via, dev, table } => {
            add_route(&destination, via.as_deref(), dev.as_deref(), table.as_deref(), options).await
        }
        RouteCommands::Delete { destination, table } => {
            delete_route(&destination, table.as_deref(), options).await
        }
    }
}

async fn show_routes(route_type: &str, options: &GlobalOptions) -> Result<()> {
    match route_type {
        "table" => {
            log_info("Routing table:");
            println!();
            
            #[cfg(target_os = "linux")]
            {
                let output = execute_command("ip", &["route", "show", "table", "main"], options).await?;
                println!("{}", output);
            }
            
            #[cfg(target_os = "windows")]
            {
                let output = execute_command("route", &["print", "-4"], options).await?;
                println!("{}", output);
            }
            
            #[cfg(target_os = "macos")]
            {
                let output = execute_command("netstat", &["-nr", "-f", "inet"], options).await?;
                println!("{}", output);
            }
        }
        "cache" => {
            log_info("Route cache:");
            println!();
            
            #[cfg(target_os = "linux")]
            {
                let output = execute_command("ip", &["route", "show", "cache"], options).await?;
                println!("{}", output);
            }
            
            #[cfg(not(target_os = "linux"))]
            {
                log_info("Route cache display not supported on this platform");
            }
        }
        "all" => {
            log_info("All routing tables:");
            println!();
            
            #[cfg(target_os = "linux")]
            {
                let output = execute_command("ip", &["route", "show", "table", "all"], options).await?;
                println!("{}", output);
            }
            
            #[cfg(target_os = "windows")]
            {
                let output = execute_command("route", &["print"], options).await?;
                println!("{}", output);
            }
            
            #[cfg(target_os = "macos")]
            {
                let output = execute_command("netstat", &["-nr"], options).await?;
                println!("{}", output);
            }
        }
        _ => {
            log_info(&format!("Routes for {}:", route_type));
            println!();
            
            #[cfg(target_os = "linux")]
            {
                let output = execute_command("ip", &["route", "show", route_type], options).await?;
                println!("{}", output);
            }
            
            #[cfg(not(target_os = "linux"))]
            {
                return Err(NetMgrError::UnsupportedPlatform("Showing specific routes not implemented for this platform".to_string()).into());
            }
        }
    }
    
    Ok(())
}

async fn add_route(destination: &str, via: Option<&str>, dev: Option<&str>, table: Option<&str>, options: &GlobalOptions) -> Result<()> {
    log_info(&format!("Adding route: {}", destination));
    
    #[cfg(target_os = "linux")]
    {
        let mut args = vec!["route", "add", destination];
        if let Some(gateway) = via {
            args.extend(&["via", gateway]);
        }
        if let Some(device) = dev {
            args.extend(&["dev", device]);
        }
        if let Some(tbl) = table {
            if tbl != "main" {
                args.extend(&["table", tbl]);
            }
        }
        
        execute_command("ip", &args, options).await?;
    }
    
    #[cfg(target_os = "windows")]
    {
        let mut args = vec!["add", destination];
        if let Some(gateway) = via {
            args.extend(&["mask", "255.255.255.0", gateway]);
        }
        if let Some(device) = dev {
            args.extend(&["if", device]);
        }
        
        execute_command("route", &args, options).await?;
    }
    
    #[cfg(target_os = "macos")]
    {
        let mut args = vec!["add", "-net", destination];
        if let Some(gateway) = via {
            args.push(gateway);
        }
        
        execute_command("route", &args, options).await?;
    }
    
    Ok(())
}

async fn delete_route(destination: &str, table: Option<&str>, options: &GlobalOptions) -> Result<()> {
    log_info(&format!("Deleting route: {}", destination));
    
    #[cfg(target_os = "linux")]
    {
        let mut args = vec!["route", "del", destination];
        if let Some(tbl) = table {
            if tbl != "main" {
                args.extend(&["table", tbl]);
            }
        }
        
        execute_command("ip", &args, options).await?;
    }
    
    #[cfg(target_os = "windows")]
    {
        execute_command("route", &["delete", destination], options).await?;
    }
    
    #[cfg(target_os = "macos")]
    {
        execute_command("route", &["delete", destination], options).await?;
    }
    
    Ok(())
}
