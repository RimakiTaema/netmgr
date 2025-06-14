#pragma once
#include "../cli/cli.hpp"

class Diagnostic {
public:
    static int handle_command(const GlobalOptions& options);
    
private:
    static int test_connectivity(const GlobalOptions& options);
    static int test_ports(const GlobalOptions& options);
    static int monitor_bandwidth(const GlobalOptions& options);
};
