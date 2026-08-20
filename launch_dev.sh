#!/bin/bash

# launch_dev.sh
# Script to start both the Vue.js frontend and Go backend development servers.

# --- Local LLM toggle ---
# true  -> chatbot/who-to-sign/scout-report talk to the local Ollama "ornith" model
#          (Ornith-1.0-35B-MTP-APEX, see ~/models/ornith). No API key or internet
#          needed; requires Ollama installed with the model already created
#          (ollama create ornith -f ~/models/ornith/Modelfile).
# false -> normal behavior, talks to real OpenAI (gpt-5.6-luna / "Luna").
USE_LOCAL_LLM=false

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

    if [ "$OBSERVABILITY_STARTED" = "true" ]; then
        echo "Stopping local observability stack..."
        (cd observability && docker compose stop) >/dev/null 2>&1
    fi

    echo "Cleanup complete."
    exit 0
}

# Trap SIGINT (Ctrl+C) and SIGTERM to run the cleanup function
trap cleanup SIGINT SIGTERM

# --- Environment Configuration ---
# Enable protobuf serialization for improved performance
export USE_PROTOBUF=true
export MAX_UPLOAD_SIZE=100
export FORMAT_AWARE_CACHE_ENABLED=true
export S3_BUCKET_NAME=v2fmdash

# --- Extract S3 credentials from Kubernetes ---
echo "Extracting S3 credentials from Kubernetes secret 'v2fmdash-minio-secret'..."
if command -v kubectl &>/dev/null; then
    S3_ENDPOINT=$(kubectl get secret v2fmdash-minio-secret -o jsonpath='{.data.endpoint}'   2>/dev/null | base64 --decode)
    S3_ACCESS_KEY=$(kubectl get secret v2fmdash-minio-secret -o jsonpath='{.data.access-key}' 2>/dev/null | base64 --decode)
    S3_SECRET_KEY=$(kubectl get secret v2fmdash-minio-secret -o jsonpath='{.data.secret-key}' 2>/dev/null | base64 --decode)
    S3_USE_SSL=$(kubectl get secret v2fmdash-minio-secret -o jsonpath='{.data.use-ssl}'     2>/dev/null | base64 --decode)
    S3_USE_SSL="${S3_USE_SSL:-true}"

    [ -n "$S3_ENDPOINT" ]   && echo "  endpoint:   OK" || echo "  endpoint:   MISSING"
    [ -n "$S3_ACCESS_KEY" ] && echo "  access-key: OK" || echo "  access-key: MISSING"
    [ -n "$S3_SECRET_KEY" ] && echo "  secret-key: OK" || echo "  secret-key: MISSING"

    if [ -n "$S3_ENDPOINT" ] && [ -n "$S3_ACCESS_KEY" ] && [ -n "$S3_SECRET_KEY" ]; then
        export S3_ENDPOINT S3_ACCESS_KEY S3_SECRET_KEY S3_USE_SSL
        echo "S3 credentials loaded successfully."
    else
        echo "Warning: One or more S3 credentials are missing (see above). Images may not load."
    fi
else
    echo "Warning: kubectl not found. S3 credentials not loaded. Images may not load."
fi

echo "Protobuf serialization: ENABLED"
echo "Format-aware caching: ENABLED"
echo "To disable protobuf, set USE_PROTOBUF=false or comment out the export line above"
echo "To disable format-aware caching, set FORMAT_AWARE_CACHE_ENABLED=false or comment out the export line above"
echo ""

# --- Start the local OTEL observability stack: Jaeger (traces) + otel-collector +
# Prometheus + Grafana (metrics) -- tracing map ticket 06, widened after live testing
# showed Jaeger alone can't display any of the fm24_*/gen_ai.* metrics (its OTLP
# receiver only implements the trace service). The app's single OTLP endpoint now
# points at the collector, which fans traces out to Jaeger and metrics out to
# Prometheus/Grafana. Always-on by default: warn and continue without it if Docker
# isn't available, mirroring the kubectl/S3-credentials fallback above.
OBSERVABILITY_STARTED=false
if command -v docker &>/dev/null && docker info >/dev/null 2>&1; then
    echo "Starting local observability stack (Jaeger + otel-collector + Prometheus + Grafana)..."
    if (cd observability && docker compose ps --status running -q) | grep -q .; then
        echo "  Already running, reusing it."
        OBSERVABILITY_STARTED=true
    else
        (cd observability && docker compose up -d) >/dev/null 2>&1 && OBSERVABILITY_STARTED=true
    fi

    if [ "$OBSERVABILITY_STARTED" = "true" ]; then
        # Brief readiness wait (~5s) so early backend spans aren't dropped -- proceed
        # regardless once the timeout elapses, never block dev startup indefinitely.
        for i in $(seq 1 25); do
            (exec 3<>/dev/tcp/localhost/4317) >/dev/null 2>&1 && exec 3<&- 3>&- && break
            sleep 0.2
        done
        export OTEL_ENABLED=true
        export OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317
        export OTEL_EXPORTER_OTLP_INSECURE=true
        echo "  Jaeger UI:     http://localhost:16686"
        echo "  Grafana UI:    http://localhost:3001  (Prometheus + Jaeger datasources preconfigured)"
        echo "  Prometheus UI: http://localhost:9090"
    else
        echo "  Warning: could not start the observability stack. Continuing without local tracing."
    fi
else
    echo "Warning: Docker not found or not running. Continuing without local tracing."
fi
echo ""

# --- Start local LLM (Ollama) if USE_LOCAL_LLM is enabled above ---
# Mirrors the docker/kubectl fallback style: never blocks dev startup, just warns and
# falls back to gpt-5.6-luna if Ollama or the "ornith" model aren't available.
#
# Uses the HTTP API (not `ollama list`/`ollama ps`) for every readiness/model check: the
# `ollama` CLI's own client can block far longer than its process has any business taking
# when the daemon is down -- observed hanging 60s+ on this machine -- which would stall
# dev startup instead of failing over. A closed port here also doesn't fail the connect
# fast (looks like Windows Firewall blackholing rather than RST'ing it), so each failed
# HTTP attempt can eat its full timeout too; per-attempt timeouts and attempt counts below
# are kept short/bounded so a genuinely stuck Ollama still fails over well under a minute.
LOCAL_LLM_READY=false
if [ "$USE_LOCAL_LLM" = "true" ]; then
    echo "Local LLM enabled -- checking Ollama (model: ornith)..."
    if command -v ollama &>/dev/null; then
        ollama_reachable() { curl -s -m 1 http://localhost:11434/api/version >/dev/null 2>&1; }

        OLLAMA_UP=false
        ollama_reachable && OLLAMA_UP=true

        if [ "$OLLAMA_UP" = "false" ]; then
            echo "  Starting Ollama server..."
            nohup ollama serve >/dev/null 2>&1 &
            for i in $(seq 1 15); do
                ollama_reachable && { OLLAMA_UP=true; break; }
                sleep 0.3
            done
        fi

        if [ "$OLLAMA_UP" = "true" ]; then
            if curl -s -m 2 http://localhost:11434/api/tags 2>/dev/null | grep -q '"name":"ornith'; then
                LOCAL_LLM_READY=true
                export ENVIRONMENT=development
                export LOCAL_LLM_BASE_URL=http://localhost:11434/v1
                echo "  Using local LLM: http://localhost:11434/v1 (model: ornith)"
            else
                echo "  Warning: Ollama is running but the 'ornith' model was not found. Falling back to gpt-5.6-luna. Run: ollama create ornith -f ~/models/ornith/Modelfile"
            fi
        else
            echo "  Warning: Ollama did not come up in time. Falling back to gpt-5.6-luna."
        fi
    else
        echo "  Warning: Ollama not found on PATH. Falling back to gpt-5.6-luna."
    fi
    # Deliberately not stopped in cleanup() -- Ollama is a shared system service (other
    # models/tools may depend on it), unlike the fmdash-only observability stack.
else
    echo "Local LLM disabled (USE_LOCAL_LLM=false) -- using gpt-5.6-luna."
fi
echo ""

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
    (cd "$GO_API_DIR" && USE_PROTOBUF="$USE_PROTOBUF" MAX_UPLOAD_SIZE="$MAX_UPLOAD_SIZE" go run .) &
    BACKEND_PID=$! # Get the Process ID of the backgrounded Go server

    sleep 2 # Give it a moment to start or fail
    if ! ps -p $BACKEND_PID > /dev/null; then
        echo "Error: Go backend server (go run .) failed to start from $GO_API_DIR."
        BACKEND_PID="" # Clear PID if it failed
    else
        echo "Backend server process started with PID: $BACKEND_PID"
        echo "Go API should be available at http://localhost:8091 (usually proxied by Vite)"
        if [ "$LOCAL_LLM_READY" = "true" ]; then
            echo "LLM features: local (ornith via Ollama)"
        else
            echo "LLM features: gpt-5.6-luna (OpenAI)"
        fi
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
