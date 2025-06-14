#include "route.hpp"
#include "../common/common.hpp"
#include <iostream>

int Route::handle_command(const GlobalOptions& options) {
    switch (options.subcommand) {
        case SubCommandType::SHOW:
            return show_routes(options);
        case SubCommandType::ADD:
            return add_route(options);
        case SubCommandType::DELETE:
            return delete_route(options);
        default:
            std::cerr << "Unknown route subcommand" << std::endl;
            return 1;
    }
}

int Route::show_routes(const GlobalOptions& options) {
    Common::log_info("Routing table:");
    std::cout << std::endl;
    
    #ifdef __linux__
    return Common::execute_command("ip", {"route", "show"}, options.dry_run);
    #elif defined(__APPLE__)
    return Common::execute_command("netstat", {"-nr", "-f", "inet"}, options.dry_run);
    #elif defined(_WIN32)
    return Common::execute_command("route", {"print", "-4"}, options.dry_run);
    #endif
    
    return 0;
}

int Route::add_route(const GlobalOptions& options) {
    if (options.args.size() < 1) {
        std::cerr << "Usage: netmgr route add <destination> [--via gateway] [--dev interface]" << std::endl;
        return 1;
    }
    
    std::string destination = options.args[0];
    std::string gateway;
    std::string interface;
    
    // Parse additional arguments
    for (size_t i = 1; i < options.args.size(); i++) {
        if (options.args[i] == "--via" && i + 1 < options.args.size()) {
            gateway = options.args[++i];
        } else if (options.args[i] == "--dev" && i + 1 < options.args.size()) {
            interface = options.args[++i];
        }
    }
    
    Common::log_info("Adding route: " + destination);
    
    #ifdef __linux__
    std::vector<std::string> cmd_args = {"route", "add", destination};
    if (!gateway.empty()) {
        cmd_args.push_back("via");
        cmd_args.push_back(gateway);
    }
    if (!interface.empty()) {
        cmd_args.push_back("dev");
        cmd_args.push_back(interface);
    }
    return Common::execute_command("ip", cmd_args, options.dry_run);
    #elif defined(__APPLE__)
    std::vector<std::string> cmd_args = {"add", "-net", destination};
    if (!gateway.empty()) {
        cmd_args.push_back(gateway);
    }
    return Common::execute_command("route", cmd_args, options.dry_run);
    #elif defined(_WIN32)
    std::vector<std::string> cmd_args = {"add", destination};
    if (!gateway.empty()) {
        cmd_args.push_back("mask");
        cmd_args.push_back("255.255.255.0");
        cmd_args.push_back(gateway);
    }
    return Common::execute_command("route", cmd_args, options.dry_run);
    #endif
    
    return 0;
}

int Route::delete_route(const GlobalOptions& options) {
    if (options.args.empty()) {
        std::cerr << "Usage: netmgr route delete <destination>" << std::endl;
        return 1;
    }
    
    std::string destination = options.args[0];
    Common::log_info("Deleting route: " + destination);
    
    #ifdef __linux__
    return Common::execute_command("ip", {"route", "del", destination}, options.dry_run);
    #elif defined(__APPLE__)
    return Common::execute_command("route", {"delete", destination}, options.dry_run);
    #elif defined(_WIN32)
    return Common::execute_command("route", {"delete", destination}, options.dry_run);
    #endif
    
    return 0;
}
