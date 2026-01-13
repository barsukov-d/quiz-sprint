#!/bin/bash

# SSH Tunnel script for development
# Forwards local Vue dev server (port 5173) to VPS
# Run this on your MacBook while developing

VPS_HOST="144.31.199.226"
VPS_USER="root"
LOCAL_PORT="5173"
REMOTE_PORT="5173"

echo "=== Starting SSH tunnel for TMA development ==="
echo "Local:  localhost:${LOCAL_PORT}"
echo "Remote: ${VPS_HOST}:${REMOTE_PORT}"
echo "Public: https://dev.quiz-sprint-tma.online"
echo ""
echo "Press Ctrl+C to stop the tunnel"
echo ""

# -N: Don't execute remote command
# -R: Remote port forwarding (VPS port 5173 -> MacBook port 5173)
# -o ServerAliveInterval=60: Keep connection alive
# -o ServerAliveCountMax=3: Retry 3 times before giving up

ssh -N \
    -R ${REMOTE_PORT}:localhost:${LOCAL_PORT} \
    -o ServerAliveInterval=60 \
    -o ServerAliveCountMax=3 \
    ${VPS_USER}@${VPS_HOST}
