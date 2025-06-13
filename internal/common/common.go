package common

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Global variables
var (
	ConfigDir string
	StateDir  string
	LogFile   string
	Opts      *Options
)

// Options represents global command options
type Options struct {
	Verbose bool
	DryRun  bool
	Force   bool
}

// Colors for terminal output
const (
	ColorRed     = "\033[0;31m"
	ColorGreen   = "\033[0;32m"
	ColorYellow  = "\033[1;33m"
	ColorBlue    = "\033[0;34m"
	ColorPurple  = "\033[0;35m"
	ColorCyan    = "\033[0;36m"
	ColorWhite   = "\033[1;37m"
	ColorGray    = "\033[0;37m"
	ColorReset   = "\033[0m"
	NoColor      = ""
)

// SetOptions sets the global options
func SetOptions(options *Options) {
	Opts = options
}

// InitSystem initializes the system directories and files
func InitSystem() error {
	// Determine config and state directories based on OS
	if IsWindows() {
		ConfigDir = filepath.Join(os.Getenv("PROGRAMDATA"), "NetMgr")
		StateDir = filepath.Join(os.Getenv("PROGRAMDATA"), "NetMgr", "state")
		LogFile = filepath.Join(os.Getenv("PROGRAMDATA"), "NetMgr", "logs", "netmgr.log")
	} else {
		ConfigDir = "/etc/netmgr"
		StateDir = "/var/lib/netmgr"
		LogFile = "/var/log/netmgr.log"
	}

	// Create directories
	if err := os.MkdirAll(ConfigDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}
	if err := os.MkdirAll(StateDir, 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(LogFile), 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %v", err)
	}

	// Create default config files if they don't exist
	defaultConfigs := []string{
		"interfaces.json",
		"forwarding.json",
		"firewall.json",
		"routing.json",
		"dns.json",
		"tunnels.json",
	}

	for _, config := range defaultConfigs {
		configPath := filepath.Join(ConfigDir, config)
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			if err := ioutil.WriteFile(configPath, []byte("{}"), 0644); err != nil {
				return fmt.Errorf("failed to create config file %s: %v", config, err)
			}
		}
	}

	return nil
}

// IsRoot checks if the program is running with root/admin privileges
func IsRoot() bool {
	if IsWindows() {
		// On Windows, check if we can write to a protected directory
		testFile := filepath.Join(os.Getenv("WINDIR"), "temp_netmgr_test")
		f, err := os.Create(testFile)
		if err == nil {
			f.Close()
			os.Remove(testFile)
			return true
		}
		return false
	} else {
		// On Unix systems, check effective user ID
		return os.Geteuid() == 0
	}
}

// IsWindows returns true if running on Windows
func IsWindows() bool {
	return runtime.GOOS == "windows"
}

// IsMacOS returns true if running on macOS
func IsMacOS() bool {
	return runtime.GOOS == "darwin"
}

// IsLinux returns true if running on Linux
func IsLinux() bool {
	return runtime.GOOS == "linux"
}

// CheckDependencies checks if required tools are available
func CheckDependencies() error {
	var requiredTools []string

	if IsWindows() {
		requiredTools = []string{"netsh", "powershell"}
	} else if IsMacOS() {
		requiredTools = []string{"networksetup", "scutil", "pfctl"}
	} else {
		requiredTools = []string{"ip", "iptables", "tc"}
	}

	for _, tool := range requiredTools {
		if _, err := exec.LookPath(tool); err != nil {
			return fmt.Errorf("required tool not found: %s", tool)
		}
	}

	return nil
}

// Execute runs a command with dry-run support
func Execute(command string, args ...string) (string, error) {
	cmdStr := fmt.Sprintf("%s %s", command, strings.Join(args, " "))
	LogDebug("Executing: %s", cmdStr)

	if Opts.DryRun {
		LogInfo("[DRY-RUN] Would execute: %s", cmdStr)
		return "", nil
	}

	cmd := exec.Command(command, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		LogError("Command failed: %s", cmdStr)
		LogError("Error: %v", err)
		LogError("Output: %s", string(output))
		return string(output), err
	}

	LogDebug("Command successful")
	return string(output), nil
}

// Log levels
const (
	LogLevelError = "ERROR"
	LogLevelWarn  = "WARN"
	LogLevelInfo  = "INFO"
	LogLevelDebug = "DEBUG"
)

// LogMessage logs a message with the specified level
func LogMessage(level, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logEntry := fmt.Sprintf("[%s] [%s] %s\n", timestamp, level, message)

	// Write to log file
	f, err := os.OpenFile(LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		defer f.Close()
		f.WriteString(logEntry)
	}

	// Determine color based on level
	var color string
	switch level {
	case LogLevelError:
		color = ColorRed
	case LogLevelWarn:
		color = ColorYellow
	case LogLevelInfo:
		color = ColorGreen
	case LogLevelDebug:
		color = ColorGray
	default:
		color = ColorReset
	}

	// Only print debug messages if verbose mode is enabled
	if level == LogLevelDebug && !Opts.Verbose {
		return
	}

	// Print to console with color
	if color != NoColor {
		fmt.Printf("%s[%s]%s %s\n", color, level, ColorReset, message)
	} else {
		fmt.Printf("[%s] %s\n", level, message)
	}
}

// LogError logs an error message
func LogError(format string, args ...interface{}) {
	LogMessage(LogLevelError, format, args...)
}

// LogWarn logs a warning message
func LogWarn(format string, args ...interface{}) {
	LogMessage(LogLevelWarn, format, args...)
}

// LogInfo logs an info message
func LogInfo(format string, args ...interface{}) {
	LogMessage(LogLevelInfo, format, args...)
}

// LogDebug logs a debug message
func LogDebug(format string, args ...interface{}) {
	LogMessage(LogLevelDebug, format, args...)
}

// SaveConfig saves data to a config file
func SaveConfig(filename string, data interface{}) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	configPath := filepath.Join(ConfigDir, filename)
	return ioutil.WriteFile(configPath, jsonData, 0644)
}

// LoadConfig loads data from a config file
func LoadConfig(filename string, data interface{}) error {
	configPath := filepath.Join(ConfigDir, filename)
	jsonData, err := ioutil.ReadFile(configPath)
	if err != nil {
		return err
	}

	return json.Unmarshal(jsonData, data)
}
