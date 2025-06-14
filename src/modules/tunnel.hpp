#pragma once
#include "../cli/cli.hpp"

class Tunnel {
public:
    static int handle_command(const GlobalOptions& options);
    
private:
    static int create_tunnel(const GlobalOptions& options);
    static int delete_tunnel(const GlobalOptions& options);
};
