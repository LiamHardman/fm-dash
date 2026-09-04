#!/bin/bash

# launch_dev.sh
# Script to start both the Vue.js frontend and Go backend development servers.

# --- Local LLM selection ---
# Interactively asks whether AI features (chatbot/who-to-sign/scout-report) should talk
# to a local Ollama model or the real remote gpt-5.6-luna (OpenAI). Non-interactive
# overrides:
#   --llm=local|remote   Skip the interactive local/remote prompt.
#   --model=<alias>      Skip the interactive model-choice prompt (implies --llm=local).
#                        Valid aliases: qwen-ridge, ling-tiny, gemma4-e4b
# See .scratch/local-llm-selector/map.md for the full design/provisioning rationale
# (replaces the old hardcoded USE_LOCAL_LLM/"ornith" setup).
model_source() {
    case "$1" in
        qwen-ridge) echo "hf.co/empero-ai/Qwen3.8-27B-Ridge-GGUF" ;;
        ling-tiny)  echo "maternion/ling-3.0-tiny:8b" ;;
        gemma4-e4b) echo "gemma4:e4b" ;;
    esac
}
model_size() {
    case "$1" in
        qwen-ridge) echo "~12.6 GB" ;;
        ling-tiny)  echo "~5.3 GB" ;;
        gemma4-e4b) echo "~9.6 GB" ;;
    esac
}
model_label() {
    case "$1" in
        qwen-ridge) echo "Qwen3.8-27B-Ridge (empero-ai, ~12.6GB, needs a beefy GPU/RAM)" ;;
        ling-tiny)  echo "Ling-3.0-tiny:8b (inclusionAI, ~5.3GB)" ;;
        gemma4-e4b) echo "Gemma 4 E4B (google, ~9.6GB)" ;;
    esac
}

LLM_MODE=""
MODEL_ALIAS=""
for arg in "$@"; do
    case "$arg" in
        --llm=local|--llm=remote)
            LLM_MODE="${arg#--llm=}"
            ;;
        --model=*)
            candidate="${arg#--model=}"
            case "$candidate" in
                qwen-ridge|ling-tiny|gemma4-e4b) MODEL_ALIAS="$candidate" ;;
                *)
                    echo "Error: unknown --model alias '$candidate'. Valid: qwen-ridge, ling-tiny, gemma4-e4b" >&2
                    exit 1
                    ;;
            esac
            ;;
        --help|-h)
            cat <<'EOF'
Usage: ./launch_dev.sh [--llm=local|remote] [--model=<alias>]

  --llm=local|remote   Skip the interactive local/remote LLM prompt.
  --model=<alias>      Skip the interactive model prompt (implies --llm=local).
                        Valid aliases: qwen-ridge, ling-tiny, gemma4-e4b

With no flags, both are asked interactively.
EOF
            exit 0
            ;;
        *)
            echo "Warning: unknown argument '$arg' ignored." >&2
            ;;
    esac
done
if [ -n "$MODEL_ALIAS" ] && [ -z "$LLM_MODE" ]; then
    LLM_MODE="local"
fi

if [ -z "$LLM_MODE" ]; then
    echo "Use a local or remote LLM for AI features?"
    echo "  1) Remote -- gpt-5.6-luna via OpenAI (default, needs OPENAI_API_KEY)"
    echo "  2) Local  -- served through Ollama, no internet/API key needed"
    read -r -p "Enter 1 or 2 [1]: " llm_choice
    case "$llm_choice" in
        2) LLM_MODE="local" ;;
        *) LLM_MODE="remote" ;;
    esac
fi

if [ "$LLM_MODE" = "local" ] && [ -z "$MODEL_ALIAS" ]; then
    echo ""
    echo "Which local model?"
    echo "  1) qwen-ridge  -- $(model_label qwen-ridge)"
    echo "  2) ling-tiny   -- $(model_label ling-tiny)"
    echo "  3) gemma4-e4b  -- $(model_label gemma4-e4b)"
    read -r -p "Enter 1-3 [3]: " model_choice
    case "$model_choice" in
        1) MODEL_ALIAS="qwen-ridge" ;;
        2) MODEL_ALIAS="ling-tiny" ;;
        *) MODEL_ALIAS="gemma4-e4b" ;;
    esac
fi
echo ""

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

# --- Provision & start local LLM (Ollama) if LLM_MODE=local ---
# Map decision (.scratch/local-llm-selector): models are auto-provisioned via Ollama on
# first use rather than requiring manual pre-creation like the old "ornith" setup did.
# A successful `ollama pull`/`create` exit code is necessary but NOT sufficient for
# qwen-ridge/ling-tiny -- their research tickets found known upstream architecture-load
# flakiness that only surfaces at run/generate time -- so every provisioning attempt ends
# with a smoke-test call before being trusted. That smoke test sends a real tool
# definition + structured-output schema, not just a bare prompt (see
# local_llm_failures.md #1): a bare "reply with the single word: ok" generate call
# doesn't exercise Ollama's grammar compiler at all and passed cleanly for qwen-ridge even
# though every real chatbot/who-to-sign/scout-report request -- which all send tool
# defs/schemas -- failed instantly with "Failed to initialize samplers: failed to parse
# grammar". Mirrors the docker/kubectl fallback style throughout: never blocks dev
# startup, just warns and falls back to gpt-5.6-luna on any failure.
#
# Uses the HTTP API (not `ollama list`/`ollama ps`) for readiness/tag checks: the `ollama`
# CLI's own client can block far longer than its process has any business taking when the
# daemon is down -- observed hanging 60s+ on this machine -- which would stall dev startup
# instead of failing over. A closed port here also doesn't fail the connect fast (looks
# like Windows Firewall blackholing rather than RST'ing it), so each failed HTTP attempt
# can eat its full timeout too; per-attempt timeouts and attempt counts below are kept
# short/bounded so a genuinely stuck Ollama still fails over well under a minute.
LOCAL_LLM_READY=false
LOCAL_LLM_MODEL_TAG=""
if [ "$LLM_MODE" = "local" ]; then
    echo "Local LLM enabled -- checking Ollama (model: $MODEL_ALIAS)..."
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
            MODEL_SOURCE=$(model_source "$MODEL_ALIAS")
            HAS_MODEL=false
            if curl -s -m 2 http://localhost:11434/api/tags 2>/dev/null | grep -q "\"name\":\"${MODEL_ALIAS}"; then
                HAS_MODEL=true
            fi

            USER_DECLINED=false
            if [ "$HAS_MODEL" = "false" ]; then
                echo "  '$MODEL_ALIAS' not found locally -- this needs a one-time download (approx. $(model_size "$MODEL_ALIAS"))."
                read -r -p "  Proceed with download? [y/N]: " confirm_dl
                case "$confirm_dl" in
                    y|Y|yes|Yes) ;;
                    *)
                        echo "  Download declined -- falling back to gpt-5.6-luna."
                        USER_DECLINED=true
                        ;;
                esac
            fi

            if [ "$HAS_MODEL" = "false" ] && [ "$USER_DECLINED" = "false" ]; then
                echo "  Pulling $MODEL_SOURCE (this can take a while for larger models)..."
                if ollama pull "$MODEL_SOURCE" && { [ "$MODEL_SOURCE" = "$MODEL_ALIAS" ] || ollama cp "$MODEL_SOURCE" "$MODEL_ALIAS"; }; then
                    HAS_MODEL=true
                else
                    echo "  Warning: failed to pull/tag '$MODEL_ALIAS'. Falling back to gpt-5.6-luna."
                fi
            fi

            if [ "$HAS_MODEL" = "true" ]; then
                echo "  Smoke-testing '$MODEL_ALIAS' (tool-calling + structured output)..."
                # Same shape every real feature sends: a function tool def plus a strict
                # json_schema Text.Format (see src/api/chatbot.go, who_to_sign.go,
                # scout_report.go). A bare generate call can't catch a grammar-compiler
                # failure like qwen-ridge's because it never asks Ollama to compile a
                # grammar in the first place.
                SMOKE_PAYLOAD='{"model":"'"$MODEL_ALIAS"'","input":"reply with the single word: ok","tools":[{"type":"function","name":"dummy_smoke_test_tool","description":"A dummy tool used only to smoke-test tool-calling support.","parameters":{"type":"object","properties":{"x":{"type":"string"}},"required":["x"]}}],"text":{"format":{"type":"json_schema","name":"smoke_test_response","strict":true,"schema":{"type":"object","properties":{"ok":{"type":"boolean"}},"required":["ok"],"additionalProperties":false}}}}'
                SMOKE_HTTP_CODE=$(curl -s -m 90 -o /dev/null -w "%{http_code}" -X POST http://127.0.0.1:11434/v1/responses -H "Content-Type: application/json" -d "$SMOKE_PAYLOAD")
                if [ "$SMOKE_HTTP_CODE" = "200" ]; then
                    LOCAL_LLM_READY=true
                    LOCAL_LLM_MODEL_TAG="$MODEL_ALIAS"
                    export ENVIRONMENT=development
                    export LOCAL_LLM_BASE_URL=http://localhost:11434/v1
                    export LOCAL_LLM_MODEL="$MODEL_ALIAS"
                    echo "  Using local LLM: http://localhost:11434/v1 (model: $MODEL_ALIAS)"
                else
                    echo "  Warning: '$MODEL_ALIAS' failed a tool-calling/structured-output smoke test (HTTP $SMOKE_HTTP_CODE; see local_llm_failures.md -- known architecture-load/grammar-compiler flakiness on some Ollama versions). Falling back to gpt-5.6-luna."
                fi
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
    echo "Local LLM disabled (remote selected) -- using gpt-5.6-luna."
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
            echo "LLM features: local ($LOCAL_LLM_MODEL_TAG via Ollama)"
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
