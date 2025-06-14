#pragma once
#include "../cli/cli.hpp"

class Interface {
public:
    static int handle_command(const GlobalOptions& options);
    
private:
    static int show_interfaces(const GlobalOptions& options);
    static int show_interface(const std::string& name, const GlobalOptions& options);
    static int set_interface(const GlobalOptions& options);
};
