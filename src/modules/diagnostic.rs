use anyhow::Result;
use std::time::Duration;
use tokio::time::sleep;

use crate::cli::DiagCommands;
use crate::common::*;

pub async fn handle_command(command: DiagCommands, options: &GlobalOptions) -> Result<()> {
    match command {
        DiagCommands::Connectivity { target, count } => {
            test_connectivity(&target, &count, options).await
        }
        DiagCommands::Ports { target, ports } => {
            test_ports(&target, &ports, options).await
        }
        DiagCommands::Bandwidth { interface, duration } => {
            monitor_bandwidth(&interface, &duration, options).await
        }
    }
}

async fn test_connectivity(target: &str, count: &str, options: &GlobalOptions) -> Result<()> {
    log_info(&format!("Testing connectivity to {}", target));
    println!();
    
    println!("{}=== Ping Test ==={}", COLORS.cyan, COLORS.reset);
    
    #[cfg(target_os = "windows")]
    {
        let output = execute_command("ping", &["-n", count, target], options).await?;
        println!("{}", output);
    }
    
    #[cfg(not(target_os = "windows"))]
    {
        let output = execute_command("ping", &["-c", count, target], options).await?;
        println!("{}", output);
    }
    
    println!("\n{}=== Traceroute ==={}", COLORS.cyan, COLORS.reset);
    
    #[cfg(target_os = "windows")]
    {
        let output = execute_command("tracert", &[target], options).await?;
        println!("{}", output);
    }
    
    #[cfg(target_os = "macos")]
    {
        let output = execute_command("traceroute", &[target], options).await?;
        println!("{}", output);
    }
    
    #[cfg(target_os = "linux")]
    {
        // Try traceroute first, fall back to tracepath
        match execute_command("traceroute", &[target], options).await {
            Ok(output) => println!("{}", output),
            Err(_) => {
                let output = execute_command("tracepath", &[target], options).await?;
                println!("{}", output);
            }
        }
    }
    
    println!("\n{}=== DNS Resolution ==={}", COLORS.cyan, COLORS.reset);
    
    let output = execute_command("nslookup", &[target], options).await?;
    println!("{}", output);
    
    Ok(())
}

async fn test_ports(target: &str, ports: &str, options: &GlobalOptions) -> Result<()> {
    log_info(&format!("Testing ports on {}: {}", target, ports));
    println!();
    
    let port_list: Vec<&str> = ports.split(',').collect();
    
    for port in port_list {
        let port = port.trim();
        
        #[cfg(target_os = "windows")]
        {
            let cmd = format!("$conn = New-Object System.Net.Sockets.TcpClient; try {{ $conn.Connect('{}', {}); Write-Host 'Open' }} catch {{ Write-Host 'Closed' }} finally {{ $conn.Close() }}", target, port);
            let output = execute_command("powershell", &["-Command", &cmd], options).await?;
            
            if output.contains("Open") {
                println!("{}✓{} Port {} is open", COLORS.green, COLORS.reset, port);
            } else {
                println!("{}✗{} Port {} is closed or filtered", COLORS.red, COLORS.reset, port);
            }
        }
        
        #[cfg(not(target_os = "windows"))]
        {
            match execute_command("nc", &["-z", "-w", "3", target, port], options).await {
                Ok(_) => println!("{}✓{} Port {} is open", COLORS.green, COLORS.reset, port),
                Err(_) => println!("{}✗{} Port {} is closed or filtered", COLORS.red, COLORS.reset, port),
            }
        }
    }
    
    Ok(())
}

async fn monitor_bandwidth(interface: &str, duration_str: &str, options: &GlobalOptions) -> Result<()> {
    let duration: u64 = duration_str.parse()
        .map_err(|_| NetMgrError::ConfigError(format!("Invalid duration: {}", duration_str)))?;
    
    log_info(&format!("Monitoring bandwidth on {} for {}s", interface, duration));
    println!();
    
    #[cfg(target_os = "linux")]
    {
        // Get initial counters
        let (rx_start, tx_start) = get_linux_interface_counters(interface, options).await?;
        
        // Wait for specified duration
        sleep(Duration::from_secs(duration)).await;
        
        // Get final counters
        let (rx_end, tx_end) = get_linux_interface_counters(interface, options).await?;
        
        // Calculate rates
        let rx_diff = rx_end - rx_start;
        let tx_diff = tx_end - tx_start;
        
        let rx_rate = rx_diff / duration as i64;
        let tx_rate = tx_diff / duration as i64;
        
        println!("RX: {}/s", format_bytes(rx_rate));
        println!("TX: {}/s", format_bytes(tx_rate));
    }
    
    #[cfg(target_os = "windows")]
    {
        let cmd = format!(r#"
$adapter = Get-NetAdapter | Where-Object {{$_.Name -eq '{}' -or $_.InterfaceDescription -like '*{}*'}} | Select-Object -First 1
$startStats = $adapter | Get-NetAdapterStatistics
Start-Sleep -Seconds {}
$endStats = $adapter | Get-NetAdapterStatistics
$rxDiff = $endStats.ReceivedBytes - $startStats.ReceivedBytes
$txDiff = $endStats.SentBytes - $startStats.SentBytes
$rxRate = $rxDiff / {}
$txRate = $txDiff / {}
"RX: " + [math]::Round($rxRate / 1KB, 2) + " KB/s"
"TX: " + [math]::Round($txRate / 1KB, 2) + " KB/s"
"#, interface, interface, duration, duration, duration);
        
        let output = execute_command("powershell", &["-Command", &cmd], options).await?;
        println!("{}", output);
    }
    
    #[cfg(target_os = "macos")]
    {
        let cmd = format!("netstat -I {} -b -w {} 2", interface, duration);
        let output = execute_command("sh", &["-c", &cmd], options).await?;
        println!("{}", output);
    }
    
    Ok(())
}

#[cfg(target_os = "linux")]
async fn get_linux_interface_counters(interface: &str, options: &GlobalOptions) -> Result<(i64, i64)> {
    let rx_data = execute_command("cat", &[&format!("/sys/class/net/{}/statistics/rx_bytes", interface)], options).await?;
    let tx_data = execute_command("cat", &[&format!("/sys/class/net/{}/statistics/tx_bytes", interface)], options).await?;
    
    let rx_bytes: i64 = rx_data.trim().parse().unwrap_or(0);
    let tx_bytes: i64 = tx_data.trim().parse().unwrap_or(0);
    
    Ok((rx_bytes, tx_bytes))
}

fn format_bytes(bytes: i64) -> String {
    const KB: i64 = 1024;
    const MB: i64 = 1024 * KB;
    const GB: i64 = 1024 * MB;
    
    if bytes < KB {
        format!("{} B", bytes)
    } else if bytes < MB {
        format!("{:.2} KB", bytes as f64 / KB as f64)
    } else if bytes < GB {
        format!("{:.2} MB", bytes as f64 / MB as f64)
    } else {
        format!("{:.2} GB", bytes as f64 / GB as f64)
    }
}
