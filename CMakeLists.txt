cmake_minimum_required(VERSION 3.10)
project(netmgr VERSION 1.0.0)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

# Platform detection
if(UNIX AND NOT APPLE)
    add_definitions(-D__linux__)
elseif(APPLE)
    add_definitions(-D__APPLE__)
elseif(WIN32)
    add_definitions(-D_WIN32)
    add_definitions(-DWIN32_LEAN_AND_MEAN)
    add_definitions(-DNOMINMAX)
endif()

# Include directories
include_directories(src)

# Source files
file(GLOB_RECURSE SOURCES "src/*.cpp")

# Create executable
add_executable(netmgr ${SOURCES})

# Compiler-specific flags
if(MSVC)
    # MSVC compiler flags
    target_compile_options(netmgr PRIVATE
        /W4          # Warning level 4
        /permissive- # Disable non-conforming code
    )
    # Link Windows libraries
    target_link_libraries(netmgr ws2_32 iphlpapi netapi32)
else()
    # GCC/Clang compiler flags
    target_compile_options(netmgr PRIVATE
        -Wall
        -Wextra
    )
    # Link pthread on Linux
    if(UNIX AND NOT APPLE)
        target_link_libraries(netmgr pthread)
    endif()
endif()

# Find and link jsoncpp if available
find_package(PkgConfig QUIET)
if(PkgConfig_FOUND)
    pkg_check_modules(JSONCPP jsoncpp)
    if(JSONCPP_FOUND)
        target_include_directories(netmgr PRIVATE ${JSONCPP_INCLUDE_DIRS})
        target_link_libraries(netmgr ${JSONCPP_LIBRARIES})
        target_compile_options(netmgr PRIVATE ${JSONCPP_CFLAGS_OTHER})
    endif()
endif()

# Windows-specific jsoncpp handling (for vcpkg)
if(WIN32 AND NOT JSONCPP_FOUND)
    find_path(JSONCPP_INCLUDE_DIR json/json.h
        PATHS ${CMAKE_PREFIX_PATH}/include
        PATH_SUFFIXES jsoncpp
    )
    find_library(JSONCPP_LIBRARY
        NAMES jsoncpp
        PATHS ${CMAKE_PREFIX_PATH}/lib
    )
    if(JSONCPP_INCLUDE_DIR AND JSONCPP_LIBRARY)
        target_include_directories(netmgr PRIVATE ${JSONCPP_INCLUDE_DIR})
        target_link_libraries(netmgr ${JSONCPP_LIBRARY})
    endif()
endif()

# Install target
install(TARGETS netmgr DESTINATION bin)

# CPack configuration for package building
set(CPACK_PACKAGE_NAME "netmgr")
set(CPACK_PACKAGE_VERSION "1.0.0")
set(CPACK_PACKAGE_DESCRIPTION_SUMMARY "Cross-platform network management tool")
set(CPACK_PACKAGE_VENDOR "Rimaki")
set(CPACK_PACKAGE_CONTACT "rimaki@rimakiproject.online")

# Debian package specific
set(CPACK_DEBIAN_PACKAGE_DEPENDS "libc6 (>= 2.17), libstdc++6 (>= 5.4.0)")
set(CPACK_DEBIAN_PACKAGE_SECTION "net")
set(CPACK_DEBIAN_PACKAGE_PRIORITY "optional")

include(CPack)
