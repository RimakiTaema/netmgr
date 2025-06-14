#include "firewall.hpp"
#include "../common/common.hpp"
#include <iostream>

int Firewall::handle_command(const GlobalOptions& options) {
    switch (options.subcommand) {
        case SubCommandType::SHOW:
            return show_rules(options);
        case SubCommandType::ADD:
            return add_rule(options);
        case SubCommandType::FLUSH:
            return flush_rules(options);
        default:
            std::cerr << "Unknown firewall subcommand" << std::endl;
            return 1;
    }
}

int Firewall::show_rules(const GlobalOptions& options) {
    Common::log_info("Firewall rules:");
    std::cout << std::endl;
    
    #ifdef __linux__
    return Common::execute_command("iptables", {"-L", "-n", "-v", "--line-numbers"}, options.dry_run);
    #elif defined(__APPLE__)
    return Common::execute_command("pfctl", {"-s", "rules"}, options.dry_run);
    #elif defined(_WIN32)
    return Common::execute_command("netsh", {"advfirewall", "firewall", "show", "rule", "name=all"}, options.dry_run);
    #endif
    
    return 0;
}

int Firewall::add_rule(const GlobalOptions& options) {
    if (options.args.size() < 3) {
        std::cerr << "Usage: netmgr firewall add <action> <port> <protocol>" << std::endl;
        return 1;
    }
    
    std::string action = options.args[0];
    std::string port = options.args[1];
    std::string protocol = options.args[2];
    
    Common::log_info("Adding firewall rule: " + action + " " + port + "/" + protocol);
    
    #ifdef __linux__
    std::string target = (action == "allow") ? "ACCEPT" : "DROP";
    return Common::execute_command("iptables", {"-A", "INPUT", "-p", protocol, "--dport", port, "-j", target}, options.dry_run);
    #elif defined(__APPLE__)
    std::string rule_action = (action == "allow") ? "pass" : "block";
    std::string rule = rule_action + " in proto " + protocol + " from any to any port " + port;
    return Common::execute_command("sh", {"-c", "echo '" + rule + "' | pfctl -a com.netmgr/rules -f -"}, options.dry_run);
    #elif defined(_WIN32)
    std::string win_action = (action == "allow") ? "allow" : "block";
    std::string rule_name = "NetMgr-" + action + "-" + protocol + "-" + port;
    return Common::execute_command("netsh", {"advfirewall", "firewall", "add", "rule", 
                                           "name=" + rule_name, "protocol=" + protocol, 
                                           "localport=" + port, "dir=in", "action=" + win_action}, options.dry_run);
    #endif
    
    return 0;
}

int Firewall::flush_rules(const GlobalOptions& options) {
    Common::log_info("Flushing firewall rules");
    
    #ifdef __linux__
    return Common::execute_command("iptables", {"-F"}, options.dry_run);
    #elif defined(__APPLE__)
    return Common::execute_command("pfctl", {"-F", "rules"}, options.dry_run);
    #elif defined(_WIN32)
    return Common::execute_command("netsh", {"advfirewall", "reset"}, options.dry_run);
    #endif
    
    return 0;
}
