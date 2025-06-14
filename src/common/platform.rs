#[cfg(target_os = "linux")]
pub const PLATFORM: &str = "linux";

#[cfg(target_os = "windows")]
pub const PLATFORM: &str = "windows";

#[cfg(target_os = "macos")]
pub const PLATFORM: &str = "macos";

#[cfg(not(any(target_os = "linux", target_os = "windows", target_os = "macos")))]
pub const PLATFORM: &str = "unknown";

pub fn is_linux() -> bool {
    cfg!(target_os = "linux")
}

pub fn is_windows() -> bool {
    cfg!(target_os = "windows")
}

pub fn is_macos() -> bool {
    cfg!(target_os = "macos")
}
