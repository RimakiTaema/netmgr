#pragma once
#include <string>
#include <vector>

enum class CommandType {
    INTERFACE,
    ROUTE,
    FIREWALL,
    FORWARD,
    DNS,
    BANDWIDTH,
    TUNNEL,
    DIAGNOSTIC
};

enum class SubCommandType {
    SHOW,
    SET,
    ADD,
    REMOVE,
    DELETE,
    FLUSH,
    SAVE,
    RESTORE
};

struct GlobalOptions {
    bool verbose = false;
    bool dry_run = false;
    bool force = false;
    CommandType command;
    SubCommandType subcommand;
    std::vector<std::string> args;
};

class CLI {
public:
    GlobalOptions parse(int argc, char* argv[]);
    void print_help();
    void print_version();
};
