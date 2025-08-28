#!/bin/bash

# Redis Sentinel Management Script
# This script helps manage the Redis Sentinel cluster

set -e

REDIS_CLI_CMD="docker exec"
MASTER_CONTAINER="course_registration_redis_master"
SLAVE1_CONTAINER="course_registration_redis_slave_1"
SLAVE2_CONTAINER="course_registration_redis_slave_2"
SENTINEL1_CONTAINER="course_registration_redis_sentinel_1"
SENTINEL2_CONTAINER="course_registration_redis_sentinel_2"
SENTINEL3_CONTAINER="course_registration_redis_sentinel_3"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_header() {
    echo -e "${BLUE}=== $1 ===${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

check_redis_status() {
    print_header "Redis Cluster Status"
    
    echo "Master Status:"
    if docker exec $MASTER_CONTAINER redis-cli ping &>/dev/null; then
        print_success "Master is running"
        echo "Master info:"
        docker exec $MASTER_CONTAINER redis-cli info replication | grep -E "(role|connected_slaves)"
    else
        print_error "Master is not responding"
    fi
    
    echo -e "\nSlave Status:"
    for slave in $SLAVE1_CONTAINER $SLAVE2_CONTAINER; do
        if docker exec $slave redis-cli ping &>/dev/null; then
            print_success "$slave is running"
            echo "$slave info:"
            docker exec $slave redis-cli info replication | grep -E "(role|master_host|master_port)"
        else
            print_error "$slave is not responding"
        fi
    done
    
    echo -e "\nSentinel Status:"
    for sentinel in $SENTINEL1_CONTAINER $SENTINEL2_CONTAINER $SENTINEL3_CONTAINER; do
        if docker exec $sentinel redis-cli -p 26379 ping &>/dev/null; then
            print_success "$sentinel is running"
        else
            print_error "$sentinel is not responding"
        fi
    done
}

check_sentinel_info() {
    print_header "Sentinel Information"
    
    echo "Master information from Sentinel:"
    docker exec $SENTINEL1_CONTAINER redis-cli -p 26379 sentinel masters | head -20
    
    echo -e "\nSlaves information from Sentinel:"
    docker exec $SENTINEL1_CONTAINER redis-cli -p 26379 sentinel slaves mymaster
    
    echo -e "\nSentinels information:"
    docker exec $SENTINEL1_CONTAINER redis-cli -p 26379 sentinel sentinels mymaster
}

test_failover() {
    print_header "Testing Manual Failover"
    print_warning "This will trigger a manual failover. Current master will become slave."
    read -p "Are you sure? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        print_warning "Triggering failover..."
        docker exec $SENTINEL1_CONTAINER redis-cli -p 26379 sentinel failover mymaster
        print_success "Failover command sent"
        echo "Waiting 10 seconds for failover to complete..."
        sleep 10
        check_redis_status
    else
        echo "Failover cancelled"
    fi
}

test_write_operations() {
    print_header "Testing Write Operations with min-slaves-to-write = 2"
    
    # Get current master
    MASTER_IP=$(docker exec $SENTINEL1_CONTAINER redis-cli -p 26379 sentinel get-master-addr-by-name mymaster | head -1)
    if [ "$MASTER_IP" = "redis-master" ]; then
        CURRENT_MASTER=$MASTER_CONTAINER
    elif [ "$MASTER_IP" = "redis-slave-1" ]; then
        CURRENT_MASTER=$SLAVE1_CONTAINER
    elif [ "$MASTER_IP" = "redis-slave-2" ]; then
        CURRENT_MASTER=$SLAVE2_CONTAINER
    else
        print_error "Could not determine current master"
        return 1
    fi
    
    echo "Current master: $CURRENT_MASTER"
    
    # Test write with all slaves available
    echo -e "\n1. Testing write with all slaves available:"
    if docker exec $CURRENT_MASTER redis-cli set test_key "test_value_$(date +%s)" &>/dev/null; then
        print_success "Write successful with all slaves available"
    else
        print_error "Write failed"
    fi
    
    # Stop one slave and test
    echo -e "\n2. Testing write with one slave stopped:"
    docker stop $SLAVE1_CONTAINER
    sleep 5
    if docker exec $CURRENT_MASTER redis-cli set test_key "test_value_$(date +%s)" &>/dev/null; then
        print_success "Write successful with one slave available"
    else
        print_error "Write failed with one slave available"
    fi
    
    # Stop second slave and test
    echo -e "\n3. Testing write with no slaves available:"
    docker stop $SLAVE2_CONTAINER
    sleep 5
    if docker exec $CURRENT_MASTER redis-cli set test_key "test_value_$(date +%s)" 2>&1 | grep -q "NOREPLICAS"; then
        print_success "Write correctly blocked - min-slaves-to-write is working!"
    else
        print_warning "Write was allowed - check min-slaves-to-write configuration"
    fi
    
    # Restart slaves
    echo -e "\n4. Restarting slaves:"
    docker start $SLAVE1_CONTAINER $SLAVE2_CONTAINER
    sleep 10
    print_success "Slaves restarted"
    
    # Test write again
    echo -e "\n5. Testing write after slave restart:"
    if docker exec $CURRENT_MASTER redis-cli set test_key "test_value_$(date +%s)" &>/dev/null; then
        print_success "Write successful after slave restart"
    else
        print_error "Write failed after slave restart"
    fi
}

monitor_logs() {
    print_header "Monitoring Redis Logs"
    echo "Press Ctrl+C to stop monitoring"
    echo "Choose container to monitor:"
    echo "1. Master"
    echo "2. Slave 1"
    echo "3. Slave 2"
    echo "4. Sentinel 1"
    echo "5. All containers"
    
    read -p "Enter choice (1-5): " choice
    
    case $choice in
        1) docker logs -f $MASTER_CONTAINER ;;
        2) docker logs -f $SLAVE1_CONTAINER ;;
        3) docker logs -f $SLAVE2_CONTAINER ;;
        4) docker logs -f $SENTINEL1_CONTAINER ;;
        5) docker-compose logs -f redis-master redis-slave-1 redis-slave-2 redis-sentinel-1 redis-sentinel-2 redis-sentinel-3 ;;
        *) print_error "Invalid choice" ;;
    esac
}

show_help() {
    echo "Redis Sentinel Management Script"
    echo ""
    echo "Usage: $0 [command]"
    echo ""
    echo "Commands:"
    echo "  status      Check Redis cluster status"
    echo "  info        Show detailed Sentinel information"
    echo "  failover    Test manual failover"
    echo "  test-write  Test write operations with min-slaves-to-write"
    echo "  logs        Monitor container logs"
    echo "  help        Show this help message"
    echo ""
    echo "If no command is provided, interactive mode will start."
}

interactive_mode() {
    while true; do
        echo ""
        print_header "Redis Sentinel Management"
        echo "1. Check cluster status"
        echo "2. Show Sentinel information"
        echo "3. Test manual failover"
        echo "4. Test write operations"
        echo "5. Monitor logs"
        echo "6. Exit"
        echo ""
        read -p "Choose an option (1-6): " choice
        
        case $choice in
            1) check_redis_status ;;
            2) check_sentinel_info ;;
            3) test_failover ;;
            4) test_write_operations ;;
            5) monitor_logs ;;
            6) print_success "Goodbye!"; exit 0 ;;
            *) print_error "Invalid choice" ;;
        esac
        
        echo ""
        read -p "Press Enter to continue..."
    done
}

# Main script logic
case "${1:-interactive}" in
    "status")
        check_redis_status
        ;;
    "info")
        check_sentinel_info
        ;;
    "failover")
        test_failover
        ;;
    "test-write")
        test_write_operations
        ;;
    "logs")
        monitor_logs
        ;;
    "help")
        show_help
        ;;
    "interactive")
        interactive_mode
        ;;
    *)
        print_error "Unknown command: $1"
        show_help
        exit 1
        ;;
esac
