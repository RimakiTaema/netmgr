#pragma once
#include "../cli/cli.hpp"

class Firewall {
public:
    static int handle_command(const GlobalOptions& options);
    
private:
    static int show_rules(const GlobalOptions& options);
    static int add_rule(const GlobalOptions& options);
    static int flush_rules(const GlobalOptions& options);
};
