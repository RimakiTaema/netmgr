#include "bandwidth.hpp"
#include "../common/common.hpp"
#include <iostream>

int Bandwidth::handle_command(const GlobalOptions& options) {
    switch (options.subcommand) {
        case SubCommandType::SHOW:
            return show_bandwidth(options);
        default:
            if (!options.args.empty() && options.args[0] == "limit") {
                return limit_bandwidth(options);
            }
            std::cerr << "Unknown bandwidth subcommand" << std::endl;
            return 1;
    }
}

int Bandwidth::show_bandwidth(const GlobalOptions& options) {
    std::string interface = options.args.empty() ? "" : options.args[0];
    
    if (interface.empty()) {
        Common::log_info("All interface bandwidth configurations:");
    } else {
        Common::log_info("Bandwidth configuration for " + interface + ":");
    }
    std::cout << std::endl;
    
    #ifdef __linux__
    if (interface.empty()) {
        return Common::execute_command("tc", {"qdisc", "show"}, options.dry_run);
    } else {
        return Common::execute_command("tc", {"qdisc", "show", "dev", interface}, options.dry_run);
    }
    #elif defined(__APPLE__)
    return Common::execute_command("ipfw", {"pipe", "show"}, options.dry_run);
    #elif defined(_WIN32)
    return Common::execute_command("powershell", {"-Command", "Get-NetQosPolicy"}, options.dry_run);
    #endif
    
    return 0;
}

int Bandwidth::limit_bandwidth(const GlobalOptions& options) {
    if (options.args.size() < 3) {
        std::cerr << "Usage: netmgr bandwidth limit <interface> <rate>" << std::endl;
        return 1;
    }
    
    std::string interface = options.args[1];
    std::string rate = options.args[2];
    
    Common::log_info("Setting bandwidth limit on " + interface + ": " + rate);
    
    #ifdef __linux__
    // Remove existing qdisc
    Common::execute_command("tc", {"qdisc", "del", "dev", interface, "root"}, true); // Ignore errors
    
    // Add new rate limiting
    return Common::execute_command("tc", {"qdisc", "add", "dev", interface, "root", "handle", "1:", 
                                         "tbf", "rate", rate, "burst", "32kbit", "latency", "400ms"}, options.dry_run);
    #elif defined(__APPLE__)
    // macOS uses ipfw for traffic shaping - simplified
    return Common::execute_command("ipfw", {"pipe", "1", "config", "bw", rate}, options.dry_run);
    #elif defined(_WIN32)
    std::string policy_name = "NetMgr-" + interface;
    return Common::execute_command("powershell", {"-Command", 
        "New-NetQosPolicy -Name '" + policy_name + "' -NetworkProfile " + interface + " -ThrottleRateActionBitsPerSecond " + rate}, options.dry_run);
    #endif
    
    return 0;
}
