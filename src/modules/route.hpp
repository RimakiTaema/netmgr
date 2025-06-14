#pragma once
#include "../cli/cli.hpp"

class Route {
public:
    static int handle_command(const GlobalOptions& options);
    
private:
    static int show_routes(const GlobalOptions& options);
    static int add_route(const GlobalOptions& options);
    static int delete_route(const GlobalOptions& options);
};
