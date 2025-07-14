#!/bin/bash

# launch_dev.sh
# Script to start both the Vue.js frontend and Go backend development servers.

# Function to clean up background processes on script exit (Ctrl+C)
cleanup() {
    echo "" # Newline for cleaner exit messages
    echo "Stopping servers..."
    if [ ! -z "$FRONTEND_PID" ] && ps -p $FRONTEND_PID > /dev/null; then
        echo "Stopping frontend server (PID: $FRONTEND_PID)..."
        # Send SIGTERM first, then SIGKILL if it doesn't stop
        kill $FRONTEND_PID
        sleep 1 # Give it a moment to shut down
        if ps -p $FRONTEND_PID > /dev/null; then
            kill -9 $FRONTEND_PID >/dev/null 2>&1
        fi
    else
        echo "Frontend server (PID: $FRONTEND_PID) already stopped or not found."
    fi

    if [ ! -z "$BACKEND_PID" ] && ps -p $BACKEND_PID > /dev/null; then
        echo "Stopping backend server (PID: $BACKEND_PID)..."
        kill $BACKEND_PID
        sleep 1
        if ps -p $BACKEND_PID > /dev/null; then
            kill -9 $BACKEND_PID >/dev/null 2>&1
        fi
    else
        echo "Backend server (PID: $BACKEND_PID) already stopped or not found."
    fi
    echo "Cleanup complete."
    exit 0
}

# Trap SIGINT (Ctrl+C) and SIGTERM to run the cleanup function
trap cleanup SIGINT SIGTERM

# --- Start Frontend Server ---
echo "Starting Vue.js frontend development server (npm run dev)..."
# This command assumes 'npm' is in your PATH and 'package.json' is in the current directory.
# The Vite server typically runs on http://localhost:3000
npm run dev &
FRONTEND_PID=$! # Get the Process ID of the backgrounded npm script

# Check if frontend started successfully (basic check, can be improved)
sleep 2 # Give it a moment to start or fail
if ! ps -p $FRONTEND_PID > /dev/null; then
    echo "Error: Frontend server (npm run dev) failed to start."
    FRONTEND_PID="" # Clear PID if it failed
else
    echo "Frontend server process started with PID: $FRONTEND_PID"
    echo "Access frontend at http://localhost:3000 (usually)"
fi
echo ""

# --- Start Backend Server ---
GO_API_DIR="src/api"
if [ ! -d "$GO_API_DIR" ]; then
    echo "Error: Go API directory '$GO_API_DIR' not found. Cannot start backend."
else
    echo "Starting Go backend API server (from $GO_API_DIR)..."
    # This command navigates to the Go API directory and runs the main.go file.
    # The Go API server typically runs on http://localhost:8091 (as per vite.config.js proxy)
    (cd "$GO_API_DIR" && go run .) &
    BACKEND_PID=$! # Get the Process ID of the backgrounded Go server

    sleep 2 # Give it a moment to start or fail
    if ! ps -p $BACKEND_PID > /dev/null; then
        echo "Error: Go backend server (go run .) failed to start from $GO_API_DIR."
        BACKEND_PID="" # Clear PID if it failed
    else
        echo "Backend server process started with PID: $BACKEND_PID"
        echo "Go API should be available at http://localhost:8091 (usually proxied by Vite)"
    fi
fi
echo ""

# --- Information for the User ---
if [ -z "$FRONTEND_PID" ] && [ -z "$BACKEND_PID" ]; then
    echo "Neither server could be started. Please check the output above for errors."
    exit 1
elif [ -z "$FRONTEND_PID" ]; then
    echo "Only the backend server appears to be starting. Frontend failed."
    echo "Press Ctrl+C to stop the backend server and this script."
elif [ -z "$BACKEND_PID" ]; then
    echo "Only the frontend server appears to be starting. Backend failed."
    echo "Press Ctrl+C to stop the frontend server and this script."
else
    echo "Both servers are attempting to run in the background."
    echo "Press Ctrl+C to stop both servers and this script."
fi

# Wait for the background processes.
# If only one PID is valid, wait for that one. If both, wait for both.
if [ ! -z "$FRONTEND_PID" ] && [ ! -z "$BACKEND_PID" ]; then
    wait $FRONTEND_PID $BACKEND_PID
elif [ ! -z "$FRONTEND_PID" ]; then
    wait $FRONTEND_PID
elif [ ! -z "$BACKEND_PID" ]; then
    wait $BACKEND_PID
fi

# If execution reaches here, it means the waited-for processes exited normally (not via Ctrl+C trap)
echo "Servers have stopped."
