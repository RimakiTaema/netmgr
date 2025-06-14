use anyhow::Result;
use serde::{Deserialize, Serialize};
use std::path::PathBuf;
use tokio::fs;

pub fn get_config_dir() -> Result<PathBuf> {
    #[cfg(windows)]
    {
        let program_data = std::env::var("PROGRAMDATA")
            .unwrap_or_else(|_| "C:\\ProgramData".to_string());
        Ok(PathBuf::from(program_data).join("NetMgr"))
    }
    
    #[cfg(unix)]
    {
        Ok(PathBuf::from("/etc/netmgr"))
    }
}

pub fn get_state_dir() -> Result<PathBuf> {
    #[cfg(windows)]
    {
        let program_data = std::env::var("PROGRAMDATA")
            .unwrap_or_else(|_| "C:\\ProgramData".to_string());
        Ok(PathBuf::from(program_data).join("NetMgr").join("state"))
    }
    
    #[cfg(unix)]
    {
        Ok(PathBuf::from("/var/lib/netmgr"))
    }
}

pub fn get_log_dir() -> Result<PathBuf> {
    #[cfg(windows)]
    {
        let program_data = std::env::var("PROGRAMDATA")
            .unwrap_or_else(|_| "C:\\ProgramData".to_string());
        Ok(PathBuf::from(program_data).join("NetMgr").join("logs"))
    }
    
    #[cfg(unix)]
    {
        Ok(PathBuf::from("/var/log"))
    }
}

pub async fn save_config<T: Serialize>(filename: &str, data: &T) -> Result<()> {
    let config_dir = get_config_dir()?;
    let config_path = config_dir.join(filename);
    let json_data = serde_json::to_string_pretty(data)?;
    fs::write(config_path, json_data).await?;
    Ok(())
}

pub async fn load_config<T: for<'de> Deserialize<'de>>(filename: &str) -> Result<T> {
    let config_dir = get_config_dir()?;
    let config_path = config_dir.join(filename);
    let json_data = fs::read_to_string(config_path).await?;
    let data = serde_json::from_str(&json_data)?;
    Ok(data)
}
