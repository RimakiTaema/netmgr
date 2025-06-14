#include "cli.hpp"
#include <iostream>
#include <algorithm>
#include <cstring>

GlobalOptions CLI::parse(int argc, char* argv[]) {
    GlobalOptions options;
    
    if (argc < 2) {
        print_help();
        exit(1);
    }
    
    // Parse global flags
    int cmd_start = 1;
    for (int i = 1; i < argc; i++) {
        if (strcmp(argv[i], "-v") == 0 || strcmp(argv[i], "--verbose") == 0) {
            options.verbose = true;
            cmd_start++;
        } else if (strcmp(argv[i], "-n") == 0 || strcmp(argv[i], "--dry-run") == 0) {
            options.dry_run = true;
            cmd_start++;
        } else if (strcmp(argv[i], "-f") == 0 || strcmp(argv[i], "--force") == 0) {
            options.force = true;
            cmd_start++;
        } else if (strcmp(argv[i], "--help") == 0 || strcmp(argv[i], "-h") == 0) {
            print_help();
            exit(0);
        } else if (strcmp(argv[i], "--version") == 0) {
            print_version();
            exit(0);
        } else {
            break;
        }
    }
    
    if (cmd_start >= argc) {
        print_help();
        exit(1);
    }
    
    // Parse command
    std::string cmd = argv[cmd_start];
    if (cmd == "interface" || cmd == "int") {
        options.command = CommandType::INTERFACE;
    } else if (cmd == "route" || cmd == "rt") {
        options.command = CommandType::ROUTE;
    } else if (cmd == "firewall" || cmd == "fw") {
        options.command = CommandType::FIREWALL;
    } else if (cmd == "forward" || cmd == "fwd") {
        options.command = CommandType::FORWARD;
    } else if (cmd == "dns") {
        options.command = CommandType::DNS;
    } else if (cmd == "bandwidth" || cmd == "bw") {
        options.command = CommandType::BANDWIDTH;
    } else if (cmd == "tunnel" || cmd == "tun") {
        options.command = CommandType::TUNNEL;
    } else if (cmd == "diagnostic" || cmd == "diag") {
        options.command = CommandType::DIAGNOSTIC;
    } else {
        std::cerr << "Unknown command: " << cmd << std::endl;
        exit(1);
    }
    
    // Parse subcommand and remaining args
    if (cmd_start + 1 < argc) {
        std::string subcmd = argv[cmd_start + 1];
        if (subcmd == "show") {
            options.subcommand = SubCommandType::SHOW;
        } else if (subcmd == "set") {
            options.subcommand = SubCommandType::SET;
        } else if (subcmd == "add") {
            options.subcommand = SubCommandType::ADD;
        } else if (subcmd == "remove") {
            options.subcommand = SubCommandType::REMOVE;
        } else if (subcmd == "delete") {
            options.subcommand = SubCommandType::DELETE;
        } else if (subcmd == "flush") {
            options.subcommand = SubCommandType::FLUSH;
        } else if (subcmd == "save") {
            options.subcommand = SubCommandType::SAVE;
        } else if (subcmd == "restore") {
            options.subcommand = SubCommandType::RESTORE;
        } else {
            options.subcommand = SubCommandType::SHOW; // default
            options.args.push_back(subcmd);
        }
        
        // Collect remaining arguments
        for (int i = cmd_start + 2; i < argc; i++) {
            options.args.push_back(argv[i]);
        }
    } else {
        options.subcommand = SubCommandType::SHOW; // default
    }
    
    return options;
}

void CLI::print_help() {
    std::cout << "netmgr - Cross-platform network management tool\n\n";
    std::cout << "USAGE:\n";
    std::cout << "    netmgr [OPTIONS] <COMMAND> [SUBCOMMAND] [ARGS...]\n\n";
    std::cout << "OPTIONS:\n";
    std::cout << "    -v, --verbose    Enable verbose output\n";
    std::cout << "    -n, --dry-run    Show what would be done without executing\n";
    std::cout << "    -f, --force      Force operations without confirmation\n";
    std::cout << "    -h, --help       Print help information\n";
    std::cout << "        --version    Print version information\n\n";
    std::cout << "COMMANDS:\n";
    std::cout << "    interface, int   Network interface management\n";
    std::cout << "    route, rt        Routing table management\n";
    std::cout << "    firewall, fw     Firewall rules management\n";
    std::cout << "    forward, fwd     Port forwarding management\n";
    std::cout << "    dns              DNS configuration\n";
    std::cout << "    bandwidth, bw    Traffic shaping and QoS\n";
    std::cout << "    tunnel, tun      Tunnel interfaces\n";
    std::cout << "    diagnostic, diag Network diagnostics\n";
}

void CLI::print_version() {
    std::cout << "netmgr 1.0.0\n";
}
