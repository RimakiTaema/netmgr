#include "diagnostic.hpp"
#include "../common/common.hpp"
#include <iostream>

int Diagnostic::handle_command(const GlobalOptions& options) {
    if (!options.args.empty()) {
        if (options.args[0] == "connectivity") {
            return test_connectivity(options);
        } else if (options.args[0] == "ports") {
            return test_ports(options);
        } else if (options.args[0] == "bandwidth") {
            return monitor_bandwidth(options);
        }
    }
    
    std::cerr << "Usage: netmgr diagnostic <connectivity|ports|bandwidth> ..." << std::endl;
    return 1;
}

int Diagnostic::test_connectivity(const GlobalOptions& options) {
    std::string target = options.args.size() > 1 ? options.args[1] : "8.8.8.8";
    std::string count = options.args.size() > 2 ? options.args[2] : "3";
    
    Common::log_info("Testing connectivity to " + target);
    std::cout << std::endl;
    
    std::cout << Colors::CYAN << "=== Ping Test ===" << Colors::RESET << std::endl;
    
    #ifdef _WIN32
    Common::execute_command("ping", {"-n", count, target}, options.dry_run);
    #else
    Common::execute_command("ping", {"-c", count, target}, options.dry_run);
    #endif
    
    std::cout << std::endl << Colors::CYAN << "=== Traceroute ===" << Colors::RESET << std::endl;
    
    #ifdef _WIN32
    Common::execute_command("tracert", {target}, options.dry_run);
    #elif defined(__APPLE__)
    Common::execute_command("traceroute", {target}, options.dry_run);
    #else
    // Try traceroute first, fall back to tracepath
    if (Common::execute_command("traceroute", {target}, options.dry_run) != 0) {
        Common::execute_command("tracepath", {target}, options.dry_run);
    }
    #endif
    
    return 0;
}

int Diagnostic::test_ports(const GlobalOptions& options) {
    if (options.args.size() < 2) {
        std::cerr << "Usage: netmgr diagnostic ports <target> [ports]" << std::endl;
        return 1;
    }
    
    std::string target = options.args[1];
    std::string ports = options.args.size() > 2 ? options.args[2] : "22,80,443";
    
    Common::log_info("Testing ports on " + target + ": " + ports);
    std::cout << std::endl;
    
    // This is simplified - would need proper port parsing and testing
    #ifdef _WIN32
    std::string cmd = "powershell -Command \"$ports = '" + ports + "'.Split(','); foreach($port in $ports) { $conn = New-Object System.Net.Sockets.TcpClient; try { $conn.Connect('" + target + "', $port); Write-Host 'Port $port is open' } catch { Write-Host 'Port $port is closed' } finally { $conn.Close() } }\"";
    return system(cmd.c_str());
    #else
    // Use netcat for port testing
    std::string cmd = "echo '" + ports + "' | tr ',' '\\n' | while read port; do if nc -z -w3 " + target + " $port 2>/dev/null; then echo \"Port $port is open\"; else echo \"Port $port is closed\"; fi; done";
    return system(cmd.c_str());
    #endif
}

int Diagnostic::monitor_bandwidth(const GlobalOptions& options) {
    std::string interface = options.args.size() > 1 ? options.args[1] : "eth0";
    std::string duration = options.args.size() > 2 ? options.args[2] : "10";
    
    Common::log_info("Monitoring bandwidth on " + interface + " for " + duration + "s");
    std::cout << std::endl;
    
    #ifdef __linux__
    std::string cmd = "sar -n DEV " + duration + " 1 | grep " + interface;
    return system(cmd.c_str());
    #elif defined(__APPLE__)
    std::string cmd = "netstat -I " + interface + " -b -w " + duration + " 2";
    return system(cmd.c_str());
    #elif defined(_WIN32)
    std::string cmd = "powershell -Command \"$adapter = Get-NetAdapter | Where-Object {$_.Name -eq '" + interface + "'} | Select-Object -First 1; $startStats = $adapter | Get-NetAdapterStatistics; Start-Sleep -Seconds " + duration + "; $endStats = $adapter | Get-NetAdapterStatistics; Write-Host ('RX: ' + [math]::Round(($endStats.ReceivedBytes - $startStats.ReceivedBytes) / " + duration + " / 1KB, 2) + ' KB/s'); Write-Host ('TX: ' + [math]::Round(($endStats.SentBytes - $startStats.SentBytes) / " + duration + " / 1KB, 2) + ' KB/s')\"";
    return system(cmd.c_str());
    #endif
    
    return 0;
}
