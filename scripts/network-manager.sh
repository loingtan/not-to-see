#!/bin/bash

# Network management script for Course Registration System
set -e

NETWORK_NAME="cobra-template_app-network"

create_network() {
    echo "🌐 Creating Docker network: $NETWORK_NAME"
    if docker network ls | grep -q "$NETWORK_NAME"; then
        echo "ℹ️  Network $NETWORK_NAME already exists"
    else
        docker network create "$NETWORK_NAME"
        echo "✅ Network $NETWORK_NAME created successfully"
    fi
}

remove_network() {
    echo "🗑️  Removing Docker network: $NETWORK_NAME"
    if docker network ls | grep -q "$NETWORK_NAME"; then
        docker network rm "$NETWORK_NAME" 2>/dev/null || echo "⚠️  Network removal failed (may be in use)"
    else
        echo "ℹ️  Network $NETWORK_NAME does not exist"
    fi
}

case "$1" in
    "create")
        create_network
        ;;
    "remove")
        remove_network
        ;;
    "recreate")
        remove_network
        sleep 2
        create_network
        ;;
    *)
        echo "Usage: $0 {create|remove|recreate}"
        echo "  create    - Create the Docker network"
        echo "  remove    - Remove the Docker network"
        echo "  recreate  - Remove and recreate the Docker network"
        exit 1
        ;;
esac
