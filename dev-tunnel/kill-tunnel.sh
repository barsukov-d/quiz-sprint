#!/bin/bash

# Kill existing SSH tunnel on VPS
# Run this if you get "remote port forwarding failed" error

VPS_HOST="144.31.199.226"
VPS_USER="root"
REMOTE_PORT="5173"

echo "=== Killing existing SSH tunnels on VPS ==="
echo "VPS: ${VPS_HOST}"
echo "Port: ${REMOTE_PORT}"
echo ""

# Find and kill process using port 5173
ssh ${VPS_USER}@${VPS_HOST} << 'EOF'
echo "Checking for processes on port 5173..."
PID=$(lsof -ti:5173 2>/dev/null || ss -lptn "sport = :5173" 2>/dev/null | grep -oP 'pid=\K[0-9]+' | head -1)

if [ -z "$PID" ]; then
    echo "No process found on port 5173"
else
    echo "Found process: $PID"
    kill -9 $PID 2>/dev/null && echo "Killed process $PID" || echo "Failed to kill process"
fi

# Also kill any lingering sshd sessions
pkill -9 -f "sshd.*5173" 2>/dev/null && echo "Killed SSH tunnels" || echo "No SSH tunnels found"

# Check if port is now free
sleep 1
if netstat -ln 2>/dev/null | grep -q ":5173 " || ss -ln 2>/dev/null | grep -q ":5173 "; then
    echo "⚠️  Port 5173 still in use!"
    netstat -ln 2>/dev/null | grep ":5173 " || ss -ln | grep ":5173 "
else
    echo "✅ Port 5173 is now free"
fi
EOF

echo ""
echo "✅ Done! You can now run ./start-dev-tunnel.sh"
