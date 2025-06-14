#pragma once
#include "../cli/cli.hpp"

class Bandwidth {
public:
    static int handle_command(const GlobalOptions& options);
    
private:
    static int show_bandwidth(const GlobalOptions& options);
    static int limit_bandwidth(const GlobalOptions& options);
};
