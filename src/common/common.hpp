#pragma once
#include <string>
#include <vector>
#include <map>

class Common {
public:
    static void init_logging(bool verbose);
    static bool is_root();
    static bool check_dependencies();
    static int execute_command(const std::string& command, const std::vector<std::string>& args, bool dry_run = false);
    static std::string execute_command_output(const std::string& command, const std::vector<std::string>& args);
    static void log_info(const std::string& message);
    static void log_error(const std::string& message);
    static void log_debug(const std::string& message);
    
private:
    static bool verbose_logging;
};

// Color constants
namespace Colors {
    extern const std::string RESET;
    extern const std::string RED;
    extern const std::string GREEN;
    extern const std::string YELLOW;
    extern const std::string BLUE;
    extern const std::string MAGENTA;
    extern const std::string CYAN;
    extern const std::string WHITE;
}
