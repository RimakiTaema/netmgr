#include "forward.hpp"
#include "../common/common.hpp"
#include <iostream>

int Forward::handle_command(const GlobalOptions& options) {
    switch (options.subcommand) {
        case SubCommandType::SHOW:
            return show_forwards(options);
        case SubCommandType::ADD:
            return add_forward(options);
        case SubCommandType::REMOVE:
            return remove_forward(options);
        default:
            std::cerr << "Unknown forward subcommand" << std::endl;
            return 1;
    }
}

int Forward::show_forwards(const GlobalOptions& options) {
    Common::log_info("Active port forwards:");
    std::cout << std::endl;
    
    #ifdef __linux__
    return Common::execute_command("iptables", {"-t", "nat", "-L", "PREROUTING", "-n", "--line-numbers"}, options.dry_run);
    #elif defined(__APPLE__)
    return Common::execute_command("pfctl", {"-s", "nat"}, options.dry_run);
    #elif defined(_WIN32)
    return Common::execute_command("netsh", {"interface", "portproxy", "show", "all"}, options.dry_run);
    #endif
    
    return 0;
}

int Forward::add_forward(const GlobalOptions& options) {
    if (options.args.size() < 4) {
        std::cerr << "Usage: netmgr forward add <name> <src_port> <dest_ip> <dest_port> [protocol]" << std::endl;
        return 1;
    }
    
    std::string name = options.args[0];
    std::string src_port = options.args[1];
    std::string dest_ip = options.args[2];
    std::string dest_port = options.args[3];
    std::string protocol = options.args.size() > 4 ? options.args[4] : "tcp";
    
    Common::log_info("Adding port forward: " + name + " (" + src_port + " -> " + dest_ip + ":" + dest_port + ")");
    
    #ifdef __linux__
    // Enable IP forwarding
    Common::execute_command("sysctl", {"-w", "net.ipv4.ip_forward=1"}, options.dry_run);
    
    // Add DNAT rule
    std::string dnat_target = "--to-destination=" + dest_ip + ":" + dest_port;
    Common::execute_command("iptables", {"-t", "nat", "-A", "PREROUTING", "-p", protocol, 
                                        "--dport", src_port, "-j", "DNAT", dnat_target}, options.dry_run);
    
    // Add FORWARD rule
    return Common::execute_command("iptables", {"-A", "FORWARD", "-p", protocol, "-d", dest_ip, 
                                               "--dport", dest_port, "-j", "ACCEPT"}, options.dry_run);
    #elif defined(__APPLE__)
    std::string rule = "rdr pass on lo0 proto " + protocol + " from any to any port " + src_port + 
                      " -> " + dest_ip + " port " + dest_port;
    return Common::execute_command("sh", {"-c", "echo '" + rule + "' | pfctl -a com.netmgr/" + name + " -f -"}, options.dry_run);
    #elif defined(_WIN32)
    return Common::execute_command("netsh", {"interface", "portproxy", "add", "v4tov4", 
                                           "listenport=" + src_port, "listenaddress=0.0.0.0",
                                           "connectport=" + dest_port, "connectaddress=" + dest_ip}, options.dry_run);
    #endif
    
    return 0;
}

int Forward::remove_forward(const GlobalOptions& options) {
    if (options.args.empty()) {
        std::cerr << "Usage: netmgr forward remove <name>" << std::endl;
        return 1;
    }
    
    std::string name = options.args[0];
    Common::log_info("Removing port forward: " + name);
    
    #ifdef __linux__
    // This is simplified - in practice you'd need to track the rules
    std::cerr << "Rule removal requires manual iptables management on Linux" << std::endl;
    return 1;
    #elif defined(__APPLE__)
    return Common::execute_command("pfctl", {"-a", "com.netmgr/" + name, "-F", "all"}, options.dry_run);
    #elif defined(_WIN32)
    // This is simplified - you'd need to track the port mappings
    std::cerr << "Forward removal requires specifying port details on Windows" << std::endl;
    return 1;
    #endif
    
    return 0;
}
