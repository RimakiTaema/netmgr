#include "interface.hpp"
#include "../common/common.hpp"
#include <iostream>
#include <iomanip>

int Interface::handle_command(const GlobalOptions& options) {
    switch (options.subcommand) {
        case SubCommandType::SHOW:
            if (options.args.empty()) {
                return show_interfaces(options);
            } else {
                return show_interface(options.args[0], options);
            }
        case SubCommandType::SET:
            return set_interface(options);
        default:
            std::cerr << "Unknown interface subcommand" << std::endl;
            return 1;
    }
}

int Interface::show_interfaces(const GlobalOptions& options) {
    Common::log_info("All network interfaces:");
    std::cout << std::endl;
    
    std::cout << std::left << std::setw(15) << "INTERFACE" 
              << std::setw(10) << "STATE" 
              << std::setw(15) << "IP ADDRESS" 
              << std::setw(20) << "MAC ADDRESS" 
              << std::setw(10) << "MTU" << std::endl;
    
    std::cout << std::left << std::setw(15) << "---------" 
              << std::setw(10) << "-----" 
              << std::setw(15) << "----------" 
              << std::setw(20) << "-----------" 
              << std::setw(10) << "---" << std::endl;
    
    #ifdef __linux__
    std::string output = Common::execute_command_output("ip", {"link", "show"});
    // Parse and display interface information
    // Implementation would parse the ip command output
    #elif defined(__APPLE__)
    std::string output = Common::execute_command_output("ifconfig", {});
    // Parse and display interface information
    #elif defined(_WIN32)
    std::string output = Common::execute_command_output("netsh", {"interface", "show", "interface"});
    // Parse and display interface information
    #endif
    
    return 0;
}

int Interface::show_interface(const std::string& name, const GlobalOptions& options) {
    Common::log_info("Interface details for: " + name);
    std::cout << std::endl;
    
    std::cout << Colors::CYAN << "=== Interface Information ===" << Colors::RESET << std::endl;
    std::cout << "Name: " << name << std::endl;
    
    #ifdef __linux__
    Common::execute_command("ip", {"addr", "show", name}, options.dry_run);
    #elif defined(__APPLE__)
    Common::execute_command("ifconfig", {name}, options.dry_run);
    #elif defined(_WIN32)
    Common::execute_command("netsh", {"interface", "ip", "show", "addresses", name}, options.dry_run);
    #endif
    
    return 0;
}

int Interface::set_interface(const GlobalOptions& options) {
    if (options.args.size() < 2) {
        std::cerr << "Usage: netmgr interface set <interface> <property> [value...]" << std::endl;
        return 1;
    }
    
    std::string interface = options.args[0];
    std::string property = options.args[1];
    
    if (property == "up") {
        #ifdef __linux__
        return Common::execute_command("ip", {"link", "set", interface, "up"}, options.dry_run);
        #elif defined(__APPLE__)
        return Common::execute_command("ifconfig", {interface, "up"}, options.dry_run);
        #elif defined(_WIN32)
        return Common::execute_command("netsh", {"interface", "set", "interface", interface, "enable"}, options.dry_run);
        #endif
    } else if (property == "down") {
        #ifdef __linux__
        return Common::execute_command("ip", {"link", "set", interface, "down"}, options.dry_run);
        #elif defined(__APPLE__)
        return Common::execute_command("ifconfig", {interface, "down"}, options.dry_run);
        #elif defined(_WIN32)
        return Common::execute_command("netsh", {"interface", "set", "interface", interface, "disable"}, options.dry_run);
        #endif
    } else if (property == "ip" && options.args.size() >= 3) {
        std::string ip = options.args[2];
        std::string prefix = options.args.size() > 3 ? options.args[3] : "24";
        
        #ifdef __linux__
        return Common::execute_command("ip", {"addr", "add", ip + "/" + prefix, "dev", interface}, options.dry_run);
        #elif defined(__APPLE__)
        return Common::execute_command("ifconfig", {interface, "inet", ip + "/" + prefix}, options.dry_run);
        #elif defined(_WIN32)
        return Common::execute_command("netsh", {"interface", "ip", "set", "address", interface, "static", ip, prefix}, options.dry_run);
        #endif
    }
    
    std::cerr << "Unknown property or insufficient arguments: " << property << std::endl;
    return 1;
}
