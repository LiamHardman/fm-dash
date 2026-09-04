# launch_dev.ps1
# Starts the Vue.js frontend and Go backend dev servers as background jobs.
# Press Enter at any time to stop all servers and exit.

# --- Local LLM selection ---
# Interactively asks whether AI features (chatbot/who-to-sign/scout-report) should talk
# to a local Ollama model or the real remote gpt-5.6-luna (OpenAI). Non-interactive
# overrides (parsed from $args below, not a formal param() block, so the flag spelling
# matches launch_dev.sh exactly):
#   --llm=local|remote   Skip the interactive local/remote prompt.
#   --model=<alias>      Skip the interactive model-choice prompt (implies --llm=local).
#                        Valid aliases: qwen-ridge, ling-tiny, gemma4-e4b
# See .scratch/local-llm-selector/map.md for the full design/provisioning rationale
# (replaces the old hardcoded $UseLocalLLM/"ornith" setup).
$ValidModelAliases = @('qwen-ridge', 'ling-tiny', 'gemma4-e4b')
$ModelLabels = @{
    'qwen-ridge' = 'Qwen3.8-27B-Ridge (empero-ai, ~12.6GB, needs a beefy GPU/RAM)'
    'ling-tiny'  = 'Ling-3.0-tiny:8b (inclusionAI, ~5.3GB)'
    'gemma4-e4b' = 'Gemma 4 E4B (google, ~9.6GB)'
}
$ModelSources = @{
    'qwen-ridge' = 'hf.co/empero-ai/Qwen3.8-27B-Ridge-GGUF'
    'ling-tiny'  = 'maternion/ling-3.0-tiny:8b'
    'gemma4-e4b' = 'gemma4:e4b'
}
$ModelSizes = @{
    'qwen-ridge' = '~12.6 GB'
    'ling-tiny'  = '~5.3 GB'
    'gemma4-e4b' = '~9.6 GB'
}

$LlmMode = $null
$ModelAlias = $null
foreach ($rawArg in $args) {
    if ($rawArg -match '^--llm=(local|remote)$') {
        $LlmMode = $Matches[1]
    } elseif ($rawArg -match '^--model=(.+)$') {
        $candidate = $Matches[1]
        if ($ValidModelAliases -notcontains $candidate) {
            Write-Error "Unknown --model alias '$candidate'. Valid: $($ValidModelAliases -join ', ')"
            exit 1
        }
        $ModelAlias = $candidate
    } elseif ($rawArg -eq '--help' -or $rawArg -eq '-h') {
        Write-Host @"
Usage: .\launch_dev.ps1 [--llm=local|remote] [--model=<alias>]

  --llm=local|remote   Skip the interactive local/remote LLM prompt.
  --model=<alias>      Skip the interactive model prompt (implies --llm=local).
                        Valid aliases: qwen-ridge, ling-tiny, gemma4-e4b

With no flags, both are asked interactively.
"@
        exit 0
    } else {
        Write-Warning "Unknown argument '$rawArg' ignored."
    }
}
if ($ModelAlias -and -not $LlmMode) { $LlmMode = 'local' }

if (-not $LlmMode) {
    Write-Host "Use a local or remote LLM for AI features?"
    Write-Host "  1) Remote -- gpt-5.6-luna via OpenAI (default, needs OPENAI_API_KEY)"
    Write-Host "  2) Local  -- served through Ollama, no internet/API key needed"
    $llmChoice = Read-Host "Enter 1 or 2 [1]"
    if ($llmChoice -eq '2') { $LlmMode = 'local' } else { $LlmMode = 'remote' }
}

if ($LlmMode -eq 'local' -and -not $ModelAlias) {
    Write-Host ""
    Write-Host "Which local model?"
    Write-Host "  1) qwen-ridge  -- $($ModelLabels['qwen-ridge'])"
    Write-Host "  2) ling-tiny   -- $($ModelLabels['ling-tiny'])"
    Write-Host "  3) gemma4-e4b  -- $($ModelLabels['gemma4-e4b'])"
    $modelChoice = Read-Host "Enter 1-3 [3]"
    switch ($modelChoice) {
        '1' { $ModelAlias = 'qwen-ridge' }
        '2' { $ModelAlias = 'ling-tiny' }
        default { $ModelAlias = 'gemma4-e4b' }
    }
}
Write-Host ""

Write-Host "Attempting to start development servers as background jobs..." -ForegroundColor Yellow
Write-Host "Output from both servers will be streamed here."
Write-Host "------------------------------------------------------------"

$frontendJob = $null
$backendJob  = $null
$scriptRootPath = $PSScriptRoot

$frontendPidFile = Join-Path $env:TEMP "fmdash_frontend_job.pid"
$backendPidFile  = Join-Path $env:TEMP "fmdash_backend_job.pid"
Remove-Item $frontendPidFile -ErrorAction SilentlyContinue
Remove-Item $backendPidFile  -ErrorAction SilentlyContinue

function Stop-DevServers {
    param(
        [System.Management.Automation.Job]$FeJob,
        [System.Management.Automation.Job]$BeJob
    )

    Write-Host ""
    Write-Host "Stopping dev servers..." -ForegroundColor Yellow

    # Kill entire process trees using the PIDs the jobs wrote on startup.
    # /T flag kills the process and all its children (node, vite, compiled Go binary, etc.)
    foreach ($pf in @($script:frontendPidFile, $script:backendPidFile)) {
        $jobPid = Get-Content $pf -ErrorAction SilentlyContinue
        if ($jobPid) {
            Write-Host "Killing process tree for PID $jobPid..."
            & taskkill /F /T /PID $jobPid *>$null
        }
    }
    Remove-Item $script:frontendPidFile -ErrorAction SilentlyContinue
    Remove-Item $script:backendPidFile  -ErrorAction SilentlyContinue

    if ($script:observabilityStarted) {
        Write-Host "Stopping local observability stack..."
        Push-Location (Join-Path $script:scriptRootPath "observability")
        docker compose stop *>$null
        Pop-Location
    }

    # Brief wait, then port-based sweep as a safety net
    Start-Sleep -Seconds 2

    foreach ($port in @(3000, 8091)) {
        $portPids = Get-NetTCPConnection -LocalPort $port -State Listen -ErrorAction SilentlyContinue |
                    Select-Object -ExpandProperty OwningProcess | Get-Unique
        foreach ($p in $portPids) {
            Write-Warning "Port $port still occupied by PID $p - force-killing..."
            Stop-Process -Id $p -Force -ErrorAction SilentlyContinue
        }
        if (-not $portPids) { Write-Host "Port $port is clear." -ForegroundColor Green }
    }

    # Clean up PowerShell job objects
    foreach ($job in @($FeJob, $BeJob)) {
        if ($job) {
            Stop-Job   -Job $job -ErrorAction SilentlyContinue
            Remove-Job -Job $job -Force -ErrorAction SilentlyContinue
        }
    }

    Write-Host "Shutdown complete." -ForegroundColor Green
}

# --- Extract S3 credentials from Kubernetes ---
Write-Host "Extracting S3 credentials from Kubernetes secret 'v2fmdash-minio-secret'..."
$s3Endpoint  = $null
$s3AccessKey = $null
$s3SecretKey = $null
$s3UseSSL    = "true"

if (Get-Command kubectl -ErrorAction SilentlyContinue) {
    try {
        function Decode-Base64Secret {
            param([Parameter(ValueFromPipeline=$true)][string]$encoded)
            process { [System.Text.Encoding]::UTF8.GetString([System.Convert]::FromBase64String($encoded.Trim())) }
        }
        $s3Endpoint  = kubectl get secret v2fmdash-minio-secret -o "jsonpath={.data.endpoint}"   | Decode-Base64Secret
        $s3AccessKey = kubectl get secret v2fmdash-minio-secret -o "jsonpath={.data.access-key}" | Decode-Base64Secret
        $s3SecretKey = kubectl get secret v2fmdash-minio-secret -o "jsonpath={.data.secret-key}" | Decode-Base64Secret
        $s3UseSslRaw = kubectl get secret v2fmdash-minio-secret -o "jsonpath={.data.use-ssl}"    | Decode-Base64Secret
        if ($s3UseSslRaw) { $s3UseSSL = $s3UseSslRaw }

        Write-Host "  endpoint:   $(if ($s3Endpoint)  { 'OK' } else { 'MISSING' })"
        Write-Host "  access-key: $(if ($s3AccessKey) { 'OK' } else { 'MISSING' })"
        Write-Host "  secret-key: $(if ($s3SecretKey) { 'OK' } else { 'MISSING' })"

        if ($s3Endpoint -and $s3AccessKey -and $s3SecretKey) {
            $env:S3_ENDPOINT    = $s3Endpoint
            $env:S3_ACCESS_KEY  = $s3AccessKey
            $env:S3_SECRET_KEY  = $s3SecretKey
            $env:S3_USE_SSL     = $s3UseSSL
            $env:S3_BUCKET_NAME = "v2fmdash"
            Write-Host "S3 credentials loaded successfully." -ForegroundColor Green
        } else {
            Write-Warning "One or more S3 credentials are missing (see above). Images may not load."
        }
    } catch {
        Write-Warning "Failed to extract S3 credentials: $($_.Exception.Message). Images may not load."
    }
} else {
    Write-Warning "kubectl not found. S3 credentials not loaded. Images may not load."
}

Write-Host ""

# --- Start the local OTEL observability stack: Jaeger (traces) + otel-collector +
# Prometheus + Grafana (metrics) -- tracing map ticket 06, widened after live testing
# showed Jaeger alone can't display any of the fm24_*/gen_ai.* metrics (its OTLP
# receiver only implements the trace service). The app's single OTLP endpoint now
# points at the collector, which fans traces out to Jaeger and metrics out to
# Prometheus/Grafana. Always-on by default: warn and continue without it if Docker
# isn't available, mirroring the kubectl/S3-credentials fallback above.
$observabilityStarted = $false
$otelEnabled = $null
$otelEndpoint = $null
$otelInsecure = $null

if (Get-Command docker -ErrorAction SilentlyContinue) {
    docker info *>$null
    if ($?) {
        Write-Host "Starting local observability stack (Jaeger + otel-collector + Prometheus + Grafana)..."
        Push-Location (Join-Path $scriptRootPath "observability")
        $running = docker compose ps --status running -q
        if ($running) {
            Write-Host "  Already running, reusing it."
            $observabilityStarted = $true
        } else {
            docker compose up -d *>$null
            $observabilityStarted = $?
        }
        Pop-Location

        if ($observabilityStarted) {
            # Brief readiness wait (~5s) so early backend spans aren't dropped -- proceed
            # regardless once the timeout elapses, never block dev startup indefinitely.
            for ($i = 0; $i -lt 25; $i++) {
                $ready = Test-NetConnection -ComputerName localhost -Port 4317 -InformationLevel Quiet -WarningAction SilentlyContinue -ErrorAction SilentlyContinue
                if ($ready) { break }
                Start-Sleep -Milliseconds 200
            }
            $otelEnabled = "true"
            $otelEndpoint = "localhost:4317"
            $otelInsecure = "true"
            Write-Host "  Jaeger UI:     http://localhost:16686" -ForegroundColor Cyan
            Write-Host "  Grafana UI:    http://localhost:3001  (Prometheus + Jaeger datasources preconfigured)" -ForegroundColor Cyan
            Write-Host "  Prometheus UI: http://localhost:9090" -ForegroundColor Cyan
        } else {
            Write-Warning "Could not start the observability stack. Continuing without local tracing."
        }
    } else {
        Write-Warning "Docker is installed but not running. Continuing without local tracing."
    }
} else {
    Write-Warning "Docker not found. Continuing without local tracing."
}
Write-Host ""

# --- Provision & start local LLM (Ollama) if $LlmMode -eq 'local' ---
# Map decision (.scratch/local-llm-selector): models are auto-provisioned via Ollama on
# first use rather than requiring manual pre-creation like the old "ornith" setup did.
# A successful `ollama pull`/`create` exit code is necessary but NOT sufficient for
# qwen-ridge/ling-tiny -- their research tickets found known upstream architecture-load
# flakiness that only surfaces at run/generate time -- so every provisioning attempt ends
# with a short smoke-test generate call before being trusted. Mirrors the docker/kubectl
# fallback style throughout: never blocks dev startup, just warns and falls back to
# gpt-5.6-luna on any failure.
#
# Uses the HTTP API (not `ollama list`/`ollama ps`) for readiness/tag checks: the `ollama`
# CLI's own client can block far longer than its process has any business taking when the
# daemon is down -- observed hanging 60s+ on this machine -- which would stall dev startup
# instead of failing over. A closed port here also doesn't fail the connect fast (looks
# like Windows Firewall blackholing rather than RST'ing it), so each failed HTTP attempt
# can eat its full timeout too; per-attempt timeouts and attempt counts below are kept
# short/bounded so a genuinely stuck Ollama still fails over well under a minute.
$localLLMReady = $false
$localLLMModelTag = $null
if ($LlmMode -eq 'local') {
    Write-Host "Local LLM enabled -- checking Ollama (model: $ModelAlias)..." -ForegroundColor Yellow
    if (Get-Command ollama -ErrorAction SilentlyContinue) {
        function Test-OllamaReachable {
            try {
                Invoke-RestMethod -Uri "http://127.0.0.1:11434/api/version" -TimeoutSec 1 -ErrorAction Stop | Out-Null
                return $true
            } catch { return $false }
        }

        $ollamaUp = Test-OllamaReachable
        if (-not $ollamaUp) {
            Write-Host "  Starting Ollama server..."
            Start-Process -FilePath "ollama" -ArgumentList "serve" -WindowStyle Hidden
            for ($i = 0; $i -lt 15; $i++) {
                if (Test-OllamaReachable) { $ollamaUp = $true; break }
                Start-Sleep -Milliseconds 300
            }
        }

        if (-not $ollamaUp) {
            Write-Warning "Ollama did not come up in time. Falling back to gpt-5.6-luna."
        } else {
            $modelSource = $ModelSources[$ModelAlias]
            $hasModel = $false
            try {
                $tags = Invoke-RestMethod -Uri "http://127.0.0.1:11434/api/tags" -TimeoutSec 2 -ErrorAction Stop
                $hasModel = [bool]($tags.models | Where-Object { $_.name -like "$ModelAlias*" })
            } catch { $hasModel = $false }

            $userDeclined = $false
            if (-not $hasModel) {
                Write-Host "  '$ModelAlias' not found locally -- this needs a one-time download (approx. $($ModelSizes[$ModelAlias]))."
                $confirmDl = Read-Host "  Proceed with download? [y/N]"
                if ($confirmDl -notmatch '^(y|yes)$') {
                    Write-Host "  Download declined -- falling back to gpt-5.6-luna."
                    $userDeclined = $true
                }
            }

            if (-not $hasModel -and -not $userDeclined) {
                Write-Host "  Pulling $modelSource (this can take a while for larger models)..."
                & ollama pull $modelSource
                $pullOk = $?
                if ($pullOk -and $modelSource -ne $ModelAlias) {
                    & ollama cp $modelSource $ModelAlias
                    $pullOk = $?
                }
                if ($pullOk) {
                    $hasModel = $true
                } else {
                    Write-Warning "Failed to pull/tag '$ModelAlias'. Falling back to gpt-5.6-luna."
                }
            }

            if ($hasModel) {
                Write-Host "  Smoke-testing '$ModelAlias'..."
                $smokeJob = Start-Job -ScriptBlock {
                    param($tag)
                    & ollama run $tag "reply with the single word: ok" *>$null
                    $LASTEXITCODE
                } -ArgumentList $ModelAlias
                $null = Wait-Job -Job $smokeJob -Timeout 60
                $smokeTimedOut = $smokeJob.State -eq 'Running'
                $smokeExit = if ($smokeTimedOut) { $null } else { Receive-Job -Job $smokeJob -ErrorAction SilentlyContinue | Select-Object -Last 1 }
                Stop-Job -Job $smokeJob -ErrorAction SilentlyContinue | Out-Null
                Remove-Job -Job $smokeJob -Force -ErrorAction SilentlyContinue

                if (-not $smokeTimedOut -and $smokeExit -eq 0) {
                    $localLLMReady = $true
                    $localLLMModelTag = $ModelAlias
                    Write-Host "  Using local LLM: http://localhost:11434/v1 (model: $ModelAlias)" -ForegroundColor Cyan
                } else {
                    Write-Warning "'$ModelAlias' failed to load/generate (see .scratch/local-llm-selector -- known architecture-load flakiness on some Ollama versions). Falling back to gpt-5.6-luna."
                }
            }
        }
    } else {
        Write-Warning "Ollama not found on PATH. Falling back to gpt-5.6-luna."
    }
    # Deliberately not stopped in Stop-DevServers -- Ollama is a shared system service
    # (other models/tools may depend on it), unlike the fmdash-only observability stack.
} else {
    Write-Host "Local LLM disabled (remote selected) -- using gpt-5.6-luna." -ForegroundColor DarkGray
}
Write-Host ""

try {
    # --- Start Frontend ---
    try {
        Write-Host "Starting Vue.js frontend (npm run dev) from '$scriptRootPath'..."
        $frontendJob = Start-Job -Name "FrontendDevServer" -ScriptBlock {
            param($path, $pidFile)
            [System.IO.File]::WriteAllText($pidFile, "$PID")
            Set-Location -Path $path
            npm run dev *>&1
        } -ArgumentList $scriptRootPath, $frontendPidFile

        if ($frontendJob) {
            Write-Host "Frontend job started (ID: $($frontendJob.Id))." -ForegroundColor Green
            Write-Host "Frontend: http://localhost:3000" -ForegroundColor Cyan
        } else { Write-Warning "Failed to start frontend job." }
    } catch { Write-Error "Exception starting frontend: $($_.Exception.Message)"; $frontendJob = $null }

    Write-Host ""

    # --- Start Backend ---
    $goApiDir = Join-Path $scriptRootPath "src/api"
    if (-not (Test-Path $goApiDir -PathType Container)) {
        Write-Error "Go API directory not found at '$goApiDir'."
        $backendJob = $null
    } else {
        try {
            Write-Host "Starting Go backend (go run . in '$goApiDir')..."
            $backendJob = Start-Job -Name "BackendDevServer" -ScriptBlock {
                param($path, $pidFile)
                [System.IO.File]::WriteAllText($pidFile, "$PID")
                Set-Location -Path $path
                $env:USE_PROTOBUF               = "true"
                $env:MAX_UPLOAD_SIZE            = "100"
                $env:FORMAT_AWARE_CACHE_ENABLED = "true"
                $env:S3_ENDPOINT               = $using:s3Endpoint
                $env:S3_ACCESS_KEY             = $using:s3AccessKey
                $env:S3_SECRET_KEY             = $using:s3SecretKey
                $env:S3_USE_SSL                = $using:s3UseSSL
                $env:S3_BUCKET_NAME            = "v2fmdash"
                if ($using:otelEnabled) {
                    $env:OTEL_ENABLED               = $using:otelEnabled
                    $env:OTEL_EXPORTER_OTLP_ENDPOINT = $using:otelEndpoint
                    $env:OTEL_EXPORTER_OTLP_INSECURE = $using:otelInsecure
                }
                if ($using:localLLMReady) {
                    $env:ENVIRONMENT       = "development"
                    $env:LOCAL_LLM_BASE_URL = "http://localhost:11434/v1"
                    $env:LOCAL_LLM_MODEL    = $using:localLLMModelTag
                }
                go run . *>&1
            } -ArgumentList $goApiDir, $backendPidFile

            if ($backendJob) {
                Write-Host "Backend job started (ID: $($backendJob.Id))." -ForegroundColor Green
                Write-Host "Backend: http://localhost:8091" -ForegroundColor Cyan
                if ($localLLMReady) {
                    Write-Host "LLM features: local ($localLLMModelTag via Ollama)" -ForegroundColor Cyan
                } else {
                    Write-Host "LLM features: gpt-5.6-luna (OpenAI)" -ForegroundColor Cyan
                }
            } else { Write-Warning "Failed to start backend job." }
        } catch { Write-Error "Exception starting backend: $($_.Exception.Message)"; $backendJob = $null }
    }

    Write-Host ""
    Write-Host "------------------------------------------------------------"

    if ($frontendJob -or $backendJob) {
        Write-Host "Both servers are running. Streaming output below." -ForegroundColor Cyan
        Write-Host "[Press Enter to stop all servers and exit]" -ForegroundColor Green
        Write-Host ""

        $shouldStop = $false
        while (-not $shouldStop) {
            if ($frontendJob -and $frontendJob.HasMoreData) {
                Receive-Job -Job $frontendJob | ForEach-Object { Write-Host "[FRONTEND] $_" }
            }
            if ($backendJob -and $backendJob.HasMoreData) {
                Receive-Job -Job $backendJob | ForEach-Object { Write-Host "[BACKEND] $_" }
            }

            # Non-blocking check for Enter key
            if ([Console]::KeyAvailable) {
                $key = [Console]::ReadKey($true)
                if ($key.Key -eq [ConsoleKey]::Enter) {
                    Write-Host "`n[Enter pressed - stopping servers...]" -ForegroundColor Yellow
                    $shouldStop = $true
                }
            }

            # Exit early if both server jobs died on their own
            $anyAlive = ($frontendJob -and $frontendJob.State -in ('Running', 'NotStarted', 'Blocked')) -or
                        ($backendJob  -and $backendJob.State  -in ('Running', 'NotStarted', 'Blocked'))
            if (-not $anyAlive) {
                Write-Host "`nAll server jobs have stopped." -ForegroundColor Yellow
                $shouldStop = $true
            }

            if (-not $shouldStop) { Start-Sleep -Milliseconds 200 }
        }
    } else {
        Write-Warning "Neither server could be started. Check errors above."
    }

} finally {
    Stop-DevServers -FeJob $frontendJob -BeJob $backendJob
}
