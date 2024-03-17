#!/bin/bash
trap "rm server;kill 0" EXIT

port_in_use() {
    local port=$1
    lsof -i :$port > /dev/null
}

for port in 8001 8002 8003 9999; do
    if port_in_use $port; then
        echo "Port $port is in use, killing the process..."
        kill $(lsof -t -i:$port)
    else
        echo "Port $port is free"
    fi
done

go build -o server
./server -port=8001 &
./server -port=8002 &
./server -port=8003 -api=1 &

wait