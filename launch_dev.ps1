# launch_dev.ps1
# Starts the Vue.js frontend and Go backend dev servers as background jobs.
# Press Enter at any time to stop all servers and exit.

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
                go run . *>&1
            } -ArgumentList $goApiDir, $backendPidFile

            if ($backendJob) {
                Write-Host "Backend job started (ID: $($backendJob.Id))." -ForegroundColor Green
                Write-Host "Backend: http://localhost:8091" -ForegroundColor Cyan
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
