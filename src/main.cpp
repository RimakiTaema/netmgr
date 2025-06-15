#include <iostream>
#include <string>
#include <vector>
#include <map>
#include <memory>
#include <cstdlib>

// Platform-specific includes
#ifndef _WIN32
#include <unistd.h>
#endif

#include "cli/cli.hpp"
#include "common/common.hpp"
#include "modules/interface.hpp"
#include "modules/route.hpp"
#include "modules/firewall.hpp"
#include "modules/forward.hpp"
#include "modules/dns.hpp"
#include "modules/bandwidth.hpp"
#include "modules/tunnel.hpp"
#include "modules/diagnostic.hpp"

int main(int argc, char* argv[]) {
    try {
        CLI cli;
        auto options = cli.parse(argc, argv);
        
        // Initialize logging
        Common::init_logging(options.verbose);
        
        // Check for root privileges on Unix systems
        #if defined(__unix__) || defined(__APPLE__)
        if (!options.dry_run && !Common::is_root()) {
            std::cerr << "This tool requires administrator privileges" << std::endl;
            return 1;
        }
        #endif
        
        // Check dependencies
        if (!Common::check_dependencies()) {
            std::cerr << "Dependency check failed" << std::endl;
            return 1;
        }
        
        // Handle commands
        switch (options.command) {
            case CommandType::INTERFACE:
                return Interface::handle_command(options);
            case CommandType::ROUTE:
                return Route::handle_command(options);
            case CommandType::FIREWALL:
                return Firewall::handle_command(options);
            case CommandType::FORWARD:
                return Forward::handle_command(options);
            case CommandType::DNS:
                return DNS::handle_command(options);
            case CommandType::BANDWIDTH:
                return Bandwidth::handle_command(options);
            case CommandType::TUNNEL:
                return Tunnel::handle_command(options);
            case CommandType::DIAGNOSTIC:
                return Diagnostic::handle_command(options);
            default:
                std::cerr << "Unknown command" << std::endl;
                return 1;
        }
        
    } catch (const std::exception& e) {
        std::cerr << "Error: " << e.what() << std::endl;
        return 1;
    }
    
    return 0;
}
