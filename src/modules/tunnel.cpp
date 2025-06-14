#include "tunnel.hpp"
#include "../common/common.hpp"
#include <iostream>

int Tunnel::handle_command(const GlobalOptions& options) {
    if (!options.args.empty() && options.args[0] == "create") {
        return create_tunnel(options);
    } else if (!options.args.empty() && options.args[0] == "delete") {
        return delete_tunnel(options);
    } else {
        std::cerr << "Usage: netmgr tunnel <create|delete> ..." << std::endl;
        return 1;
    }
}

int Tunnel::create_tunnel(const GlobalOptions& options) {
    if (options.args.size() < 5) {
        std::cerr << "Usage: netmgr tunnel create <name> <type> <local_ip> <remote_ip>" << std::endl;
        return 1;
    }
    
    std::string name = options.args[1];
    std::string type = options.args[2];
    std::string local_ip = options.args[3];
    std::string remote_ip = options.args[4];
    
    Common::log_info("Creating " + type + " tunnel: " + name);
    
    #ifdef __linux__
    int result = Common::execute_command("ip", {"tunnel", "add", name, "mode", type, 
                                               "remote", remote_ip, "local", local_ip}, options.dry_run);
    if (result == 0) {
        return Common::execute_command("ip", {"link", "set", name, "up"}, options.dry_run);
    }
    return result;
    #elif defined(__APPLE__)
    std::cerr << "Tunnel creation not implemented for macOS" << std::endl;
    return 1;
    #elif defined(_WIN32)
    if (type == "gre") {
        return Common::execute_command("netsh", {"interface", "ipv4", "add", "interface", 
                                               name, "type=tunnel", "source=" + local_ip, 
                                               "destination=" + remote_ip}, options.dry_run);
    } else {
        std::cerr << "Tunnel type " << type << " not supported on Windows" << std::endl;
        return 1;
    }
    #endif
    
    return 0;
}

int Tunnel::delete_tunnel(const GlobalOptions& options) {
    if (options.args.size() < 2) {
        std::cerr << "Usage: netmgr tunnel delete <name>" << std::endl;
        return 1;
    }
    
    std::string name = options.args[1];
    Common::log_info("Deleting tunnel: " + name);
    
    #ifdef __linux__
    Common::execute_command("ip", {"link", "set", name, "down"}, options.dry_run);
    return Common::execute_command("ip", {"tunnel", "del", name}, options.dry_run);
    #elif defined(__APPLE__)
    std::cerr << "Tunnel deletion not implemented for macOS" << std::endl;
    return 1;
    #elif defined(_WIN32)
    return Common::execute_command("netsh", {"interface", "ipv4", "delete", "interface", name}, options.dry_run);
    #endif
    
    return 0;
}
