#include "common.hpp"
#include <iostream>
#include <cstdlib>
#include <sstream>
#include <cstring>

// Platform-specific includes
#ifdef _WIN32
    #include <windows.h>
    #include <io.h>
    #include <shlobj.h>
#else
    #include <unistd.h>
#endif

bool Common::verbose_logging = false;

namespace Colors {
    const std::string RESET = "\033[0m";
    const std::string RED = "\033[31m";
    const std::string GREEN = "\033[32m";
    const std::string YELLOW = "\033[33m";
    const std::string BLUE = "\033[34m";
    const std::string MAGENTA = "\033[35m";
    const std::string CYAN = "\033[36m";
    const std::string WHITE = "\033[37m";
}

void Common::init_logging(bool verbose) {
    verbose_logging = verbose;
}

bool Common::is_root() {
    #ifdef _WIN32
    // On Windows, check if running as administrator
    BOOL isAdmin = FALSE;
    PSID administratorsGroup = NULL;
    SID_IDENTIFIER_AUTHORITY ntAuthority = SECURITY_NT_AUTHORITY;
    
    if (AllocateAndInitializeSid(&ntAuthority, 2, SECURITY_BUILTIN_DOMAIN_RID,
                                DOMAIN_ALIAS_RID_ADMINS, 0, 0, 0, 0, 0, 0,
                                &administratorsGroup)) {
        CheckTokenMembership(NULL, administratorsGroup, &isAdmin);
        FreeSid(administratorsGroup);
    }
    return isAdmin == TRUE;
    #else
    return geteuid() == 0;
    #endif
}

bool Common::check_dependencies() {
    std::vector<std::string> tools;
    
    #ifdef __linux__
    tools = {"ip", "iptables", "sysctl"};
    #elif defined(__APPLE__)
    tools = {"ifconfig", "route", "pfctl"};
    #elif defined(_WIN32)
    tools = {"netsh", "route"};
    #endif
    
    for (const auto& tool : tools) {
        std::string cmd = "which " + tool + " > /dev/null 2>&1";
        #ifdef _WIN32
        cmd = "where " + tool + " > nul 2>&1";
        #endif
        
        if (system(cmd.c_str()) != 0) {
            log_error("Required tool not found: " + tool);
            return false;
        }
    }
    
    return true;
}

int Common::execute_command(const std::string& command, const std::vector<std::string>& args, bool dry_run) {
    std::string full_cmd = command;
    for (const auto& arg : args) {
        full_cmd += " " + arg;
    }
    
    if (dry_run) {
        log_info("Would execute: " + full_cmd);
        return 0;
    }
    
    log_debug("Executing: " + full_cmd);
    return system(full_cmd.c_str());
}

std::string Common::execute_command_output(const std::string& command, const std::vector<std::string>& args) {
    std::string full_cmd = command;
    for (const auto& arg : args) {
        full_cmd += " " + arg;
    }
    
    #ifdef _WIN32
    FILE* pipe = _popen(full_cmd.c_str(), "r");
    #else
    FILE* pipe = popen(full_cmd.c_str(), "r");
    #endif
    
    if (!pipe) {
        return "";
    }
    
    std::string result;
    char buffer[128];
    while (fgets(buffer, sizeof(buffer), pipe) != nullptr) {
        result += buffer;
    }
    
    #ifdef _WIN32
    _pclose(pipe);
    #else
    pclose(pipe);
    #endif
    
    return result;
}

void Common::log_info(const std::string& message) {
    std::cout << Colors::GREEN << "[INFO] " << Colors::RESET << message << std::endl;
}

void Common::log_error(const std::string& message) {
    std::cerr << Colors::RED << "[ERROR] " << Colors::RESET << message << std::endl;
}

void Common::log_debug(const std::string& message) {
    if (verbose_logging) {
        std::cout << Colors::CYAN << "[DEBUG] " << Colors::RESET << message << std::endl;
    }
}
