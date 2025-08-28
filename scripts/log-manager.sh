#!/bin/bash

# Course Registration System Log Management Utility
# This script helps manage logging configurations for different environments

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
LOGS_DIR="$PROJECT_ROOT/logs"

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_color() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# Function to show current log configuration
show_config() {
    local config_file=$1
    print_color $BLUE "Log Configuration in $config_file:"
    if [ -f "$config_file" ]; then
        grep -A4 "^log:" "$config_file" || echo "No log configuration found"
    else
        print_color $RED "Configuration file not found: $config_file"
    fi
    echo
}

# Function to view log files
view_logs() {
    local log_file=$1
    local lines=${2:-50}
    
    if [ -f "$log_file" ]; then
        print_color $GREEN "Last $lines lines of $log_file:"
        tail -n "$lines" "$log_file"
    else
        print_color $RED "Log file not found: $log_file"
    fi
    echo
}

# Function to monitor logs in real-time
monitor_logs() {
    local log_file=$1
    
    if [ -f "$log_file" ]; then
        print_color $GREEN "Monitoring $log_file (Ctrl+C to stop):"
        tail -f "$log_file"
    else
        print_color $RED "Log file not found: $log_file"
        print_color $YELLOW "Waiting for log file to be created..."
        while [ ! -f "$log_file" ]; do
            sleep 1
        done
        print_color $GREEN "Log file created. Starting monitoring:"
        tail -f "$log_file"
    fi
}

# Function to clear logs
clear_logs() {
    if [ -d "$LOGS_DIR" ]; then
        print_color $YELLOW "Clearing all log files in $LOGS_DIR..."
        rm -f "$LOGS_DIR"/*.log
        print_color $GREEN "Log files cleared."
    else
        print_color $YELLOW "Logs directory does not exist."
    fi
}

# Function to setup log directory
setup_logs() {
    print_color $BLUE "Setting up log directory..."
    mkdir -p "$LOGS_DIR"
    chmod 755 "$LOGS_DIR"
    print_color $GREEN "Log directory created: $LOGS_DIR"
}

# Function to show usage
show_usage() {
    echo "Course Registration System Log Management Utility"
    echo ""
    echo "Usage: $0 <command> [options]"
    echo ""
    echo "Commands:"
    echo "  config <env>           Show log configuration for environment (local|development|production)"
    echo "  view <file> [lines]    View last N lines of log file (default: 50)"
    echo "  monitor <file>         Monitor log file in real-time"
    echo "  clear                  Clear all log files"
    echo "  setup                  Create logs directory"
    echo "  list                   List all log files"
    echo ""
    echo "Examples:"
    echo "  $0 config local                              # Show local config"
    echo "  $0 view logs/course-registration.log         # View last 50 lines"
    echo "  $0 view logs/course-registration.log 100     # View last 100 lines"
    echo "  $0 monitor logs/course-registration.log      # Monitor in real-time"
    echo "  $0 clear                                     # Clear all logs"
    echo "  $0 setup                                     # Setup log directory"
}

# Function to list log files
list_logs() {
    if [ -d "$LOGS_DIR" ]; then
        print_color $BLUE "Log files in $LOGS_DIR:"
        ls -la "$LOGS_DIR"/*.log 2>/dev/null || print_color $YELLOW "No log files found."
    else
        print_color $YELLOW "Logs directory does not exist. Run '$0 setup' to create it."
    fi
}

# Main script logic
case "${1:-}" in
    "config")
        ENV="${2:-local}"
        CONFIG_FILE="$PROJECT_ROOT/configs/$ENV.yaml"
        show_config "$CONFIG_FILE"
        ;;
    "view")
        LOG_FILE="${2:-$LOGS_DIR/course-registration.log}"
        LINES="${3:-50}"
        view_logs "$LOG_FILE" "$LINES"
        ;;
    "monitor")
        LOG_FILE="${2:-$LOGS_DIR/course-registration.log}"
        monitor_logs "$LOG_FILE"
        ;;
    "clear")
        clear_logs
        ;;
    "setup")
        setup_logs
        ;;
    "list")
        list_logs
        ;;
    "help"|"-h"|"--help")
        show_usage
        ;;
    *)
        print_color $RED "Invalid command: ${1:-}"
        echo
        show_usage
        exit 1
        ;;
esac
