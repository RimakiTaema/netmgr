use anyhow::Result;
use crate::cli::BandwidthCommands;
use crate::common::*;

pub async fn handle_command(command: BandwidthCommands, options: &GlobalOptions) -> Result<()> {
    match command {
        BandwidthCommands::Show { interface } => {
            show_bandwidth(interface.as_deref(), options).await
        }
        BandwidthCommands::Limit { interface, rate, burst } => {
            let burst_val = burst.as_deref().unwrap_or(&rate);
            limit_bandwidth(&interface, &rate, burst_val, options).await
        }
    }
}

async fn show_bandwidth(interface: Option<&str>, options: &GlobalOptions) -> Result<()> {
    if let Some(iface) = interface {
        log_info(&format!("Bandwidth configuration for {}:", iface));
        println!();
        
        #[cfg(target_os = "linux")]
        {
            let output = execute_command("tc", &["qdisc", "show", "dev", iface], options).await?;
            println!("{}", output);
            
            let output = execute_command("tc", &["class", "show", "dev", iface], options).await?;
            println!("{}", output);
        }
        
        #[cfg(target_os = "windows")]
        {
            let cmd = format!("Get-NetQosPolicy | Where-Object {{$_.NetworkProfile -eq '{}'}}", iface);
            let output = execute_command("powershell", &["-Command", &cmd], options).await?;
            println!("{}", output);
        }
        
        #[cfg(target_os = "macos")]
        {
            let output = execute_command("ipfw", &["pipe", "show"], options).await?;
            for line in output.lines() {
                if line.contains(iface) {
                    println!("{}", line);
                }
            }
        }
    } else {
        log_info("All interface bandwidth configurations:");
        println!();
        
        #[cfg(target_os = "linux")]
        {
            // Get all interfaces
            let output = execute_command("ip", &["link", "show"], options).await?;
            
            for line in output.lines() {
                if line.contains(": ") && !line.starts_with(' ') {
                    let parts: Vec<&str> = line.split(": ").collect();
                    if parts.len() > 1 {
                        let iface_name = parts[1].split('@').next().unwrap_or(parts[1]);
                        if iface_name != "lo" {
                            println!("\n{}=== {} ==={}", COLORS.cyan, iface_name, COLORS.reset);
                            let qdisc_output = execute_command("tc", &["qdisc", "show", "dev", iface_name], options).await?;
                            println!("{}", qdisc_output);
                        }
                    }
                }
            }
        }
        
        #[cfg(target_os = "windows")]
        {
            let output = execute_command("powershell", &["-Command", "Get-NetQosPolicy"], options).await?;
            println!("{}", output);
        }
        
        #[cfg(target_os = "macos")]
        {
            let output = execute_command("ipfw", &["pipe", "show"], options).await?;
            println!("{}", output);
        }
    }
    
    Ok(())
}

async fn limit_bandwidth(interface: &str, rate: &str, burst: &str, options: &GlobalOptions) -> Result<()> {
    log_info(&format!("Setting bandwidth limit on {}: {} (burst: {})", interface, rate, burst));
    
    #[cfg(target_os = "linux")]
    {
        // Remove existing qdisc
        let _ = execute_command("tc", &["qdisc", "del", "dev", interface, "root"], options).await;
        
        // Add new rate limiting
        execute_command("tc", &["qdisc", "add", "dev", interface, "root", "handle", "1:", 
                               "tbf", "rate", rate, "burst", burst, "latency", "70ms"], options).await?;
    }
    
    #[cfg(target_os = "windows")]
    {
        // Convert rate to bits per second
        let rate_value = parse_rate(rate);
        
        // Create QoS policy
        let policy_name = format!("NetMgr-{}", interface);
        let cmd = format!("New-NetQosPolicy -Name '{}' -NetworkProfile {} -ThrottleRateActionBitsPerSecond {}", 
                         policy_name, interface, rate_value);
        execute_command("powershell", &["-Command", &cmd], options).await?;
    }
    
    #[cfg(target_os = "macos")]
    {
        // macOS uses ipfw for traffic shaping
        let rate_value = parse_rate(rate);
        
        // Create pipe
        execute_command("ipfw", &["pipe", "1", "config", "bw", &format!("{}Kbit/s", rate_value / 1000)], options).await?;
        
        // Assign traffic to pipe
        execute_command("ipfw", &["add", "100", "pipe", "1", "ip", "from", "any", "to", "any", "via", interface], options).await?;
    }
    
    Ok(())
}

fn parse_rate(rate: &str) -> i64 {
    let rate_lower = rate.to_lowercase();
    let mut multiplier = 1i64;
    
    let rate_num = if rate_lower.ends_with("kbit") || rate_lower.ends_with("kbps") {
        multiplier = 1000;
        rate_lower.trim_end_matches("kbit").trim_end_matches("kbps")
    } else if rate_lower.ends_with("mbit") || rate_lower.ends_with("mbps") {
        multiplier = 1000000;
        rate_lower.trim_end_matches("mbit").trim_end_matches("mbps")
    } else if rate_lower.ends_with("gbit") || rate_lower.ends_with("gbps") {
        multiplier = 1000000000;
        rate_lower.trim_end_matches("gbit").trim_end_matches("gbps")
    } else {
        &rate_lower
    };
    
    rate_num.parse::<i64>().unwrap_or(0) * multiplier
}
