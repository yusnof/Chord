#!/bin/bash

# Chord DHT Test Script
# Starts multiple nodes on different ports and allows interactive testing

set -e

# Configuration
BINARY="./chord"
NODE1_PORT=4000
NODE2_PORT=4001
NODE3_PORT=4002
LOG_DIR="./logs"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Create log directory
mkdir -p "$LOG_DIR"

# Build the project
echo -e "${YELLOW}Building Chord...${NC}"
if ! go build -o "$BINARY"; then
    echo -e "${RED}Build failed!${NC}"
    exit 1
fi
echo -e "${GREEN}Build successful!${NC}"x

# Function to start a node
start_node() {
    local node_num=$1
    local port=$2
    local join_port=$3
    local is_first=$4
    
    local log_file="$LOG_DIR/node${node_num}.log"
    
    if [ "$is_first" == "true" ]; then
        echo -e "${YELLOW}Starting Node $node_num (port $port) as first node...${NC}"
        $BINARY -ipaddr "127.0.0.1" -port "$port" -ts 3000 --tff 1000 --tcp 3000 -r 4 > "$log_file" 2>&1 &
    else
        echo -e "${YELLOW}Starting Node $node_num (port $port) joining via port $join_port...${NC}"
        $BINARY -ipaddr "127.0.0.1" -port "$port" -join_addr "127.0.0.1" -join_port "$join_port" > "$log_file" 2>&1 &
    fi
    
    echo "Node $node_num PID: $!"
    sleep 2 # Give node time to start
}

# Function to stop all nodes
stop_nodes() {
    echo -e "${YELLOW}Stopping all nodes...${NC}"
    pkill -f "$BINARY" || true
    sleep 1
    echo -e "${GREEN}All nodes stopped.${NC}"
}

# Trap to cleanup on script exit
trap stop_nodes EXIT

# Main test flow
echo -e "${GREEN}=== Chord DHT Test Suite ===${NC}"
echo ""

# Option to run interactive or automated tests
if [ "$1" == "interactive" ]; then
    echo -e "${YELLOW}Starting nodes for interactive testing...${NC}"
    start_node 1 "$NODE1_PORT" "" "true"
    sleep 1
    start_node 2 "$NODE2_PORT" "$NODE1_PORT" "false"
    sleep 1
    start_node 3 "$NODE3_PORT" "$NODE1_PORT" "false"
    
    echo ""
    echo -e "${GREEN}Nodes are running!${NC}"
    echo "Node 1 logs: tail -f $LOG_DIR/node1.log"
    echo "Node 2 logs: tail -f $LOG_DIR/node2.log"
    echo "Node 3 logs: tail -f $LOG_DIR/node3.log"
    echo ""
    echo "Press Enter to stop all nodes and exit..."
    read
    
elif [ "$1" == "quick" ]; then
    echo -e "${YELLOW}Running quick test: 1 node for 5 seconds...${NC}"
    start_node 1 "$NODE1_PORT" "" "true"
    
    echo -e "${GREEN}Node running, sleeping 5 seconds...${NC}"
    sleep 100
    
    echo -e "${GREEN}Test complete!${NC}"
    echo "Check logs in $LOG_DIR/"
    
else
    echo "Usage: ./test.sh [interactive|quick]"
    echo ""
    echo "  interactive - Start 3 nodes and keep running until you press Enter"
    echo "  quick       - Start 1 node for 5 seconds"
    echo ""
    echo "Logs are saved to: $LOG_DIR/"
    exit 0
fi