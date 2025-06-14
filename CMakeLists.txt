cmake_minimum_required(VERSION 3.10)
project(netmgr VERSION 1.0.0)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

# Add compiler flags
set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -Wall -Wextra")

# Platform detection
if(UNIX AND NOT APPLE)
    add_definitions(-D__linux__)
elseif(APPLE)
    add_definitions(-D__APPLE__)
elseif(WIN32)
    add_definitions(-D_WIN32)
endif()

# Include directories
include_directories(src)

# Source files
file(GLOB_RECURSE SOURCES "src/*.cpp")

# Create executable
add_executable(netmgr ${SOURCES})

# Install target
install(TARGETS netmgr DESTINATION bin)

# CPack configuration for package building
set(CPACK_PACKAGE_NAME "netmgr")
set(CPACK_PACKAGE_VERSION "1.0.0")
set(CPACK_PACKAGE_DESCRIPTION_SUMMARY "Cross-platform network management tool")
set(CPACK_PACKAGE_VENDOR "Your Name")
set(CPACK_PACKAGE_CONTACT "your.email@example.com")

# Debian package specific
set(CPACK_DEBIAN_PACKAGE_DEPENDS "libc6 (>= 2.17), libstdc++6 (>= 5.4.0)")
set(CPACK_DEBIAN_PACKAGE_SECTION "net")
set(CPACK_DEBIAN_PACKAGE_PRIORITY "optional")

include(CPack)
