#pragma once
#include "../cli/cli.hpp"

class DNS {
public:
    static int handle_command(const GlobalOptions& options);
    
private:
    static int show_dns(const GlobalOptions& options);
    static int set_dns(const GlobalOptions& options);
};
