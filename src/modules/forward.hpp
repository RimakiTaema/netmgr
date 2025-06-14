#pragma once
#include "../cli/cli.hpp"

class Forward {
public:
    static int handle_command(const GlobalOptions& options);
    
private:
    static int show_forwards(const GlobalOptions& options);
    static int add_forward(const GlobalOptions& options);
    static int remove_forward(const GlobalOptions& options);
};
