# launch_dev.ps1
# Script to start both the Vue.js frontend and Go backend development servers as background jobs.
# Output is streamed to this window. The main script window will remain open to manage and stop the servers.

Write-Host "Attempting to start development servers as background jobs..." -ForegroundColor Yellow
Write-Host "Output from both servers will be streamed here."
Write-Host "------------------------------------------------------------"

$frontendJob = $null
$backendJob = $null
$scriptRootPath = $PSScriptRoot
$scriptInterrupted = $false # Flag to signal Ctrl+C or 's' key

# Function to clean up jobs and processes
function Stop-And-Remove-Jobs {
    param(
        [System.Management.Automation.Job]$FeJob,
        [System.Management.Automation.Job]$BeJob
    )

    Write-Host ""
    Write-Host "Initiating server shutdown and job cleanup..." -ForegroundColor Yellow

    # --- Frontend Job Cleanup (Port 3000) ---
    if ($FeJob) {
        Write-Host "Processing frontend job: $($FeJob.Name) (ID: $($FeJob.Id), Initial State: $($FeJob.State))"
        if ($FeJob.State -in ('Running', 'NotStarted', 'Blocked', 'Suspended', 'Suspending')) {
            Write-Host "Attempting to stop frontend job: $($FeJob.Name)..."
            Stop-Job -Job $FeJob -Force -ErrorAction SilentlyContinue
            Write-Host "Frontend job '$($FeJob.Name)' stop signal sent. Waiting..."
            Start-Sleep -Seconds 3
        }

        Write-Host "Checking port 3000 for lingering frontend processes..."
        $maxAttempts = 3
        for ($i = 1; $i -le $maxAttempts; $i++) {
            $processesOnPort3000 = Get-NetTCPConnection -LocalPort 3000 -State Listen -ErrorAction SilentlyContinue |
                                   Select-Object -ExpandProperty OwningProcess | Get-Unique
            if ($processesOnPort3000.Count -eq 0) {
                Write-Host "Port 3000 is clear." -ForegroundColor Green
                break
            }
            Write-Warning "Attempt $i/${maxAttempts}: Frontend process(es) still on port 3000 (PIDs: $($processesOnPort3000 -join ', ')). Terminating..."
            foreach ($pidToKill in $processesOnPort3000) {
                Write-Host "Killing PID $pidToKill on port 3000..."
                Stop-Process -Id $pidToKill -Force -ErrorAction SilentlyContinue
            }
            if ($i -lt $maxAttempts) { Start-Sleep -Seconds 2 }
            elseif ($processesOnPort3000.Count -gt 0) { Write-Error "Failed to clear port 3000 for frontend after $maxAttempts attempts." }
        }

        if ($FeJob.HasMoreData) { Write-Host "Final output from $($FeJob.Name):"; Receive-Job -Job $FeJob | ForEach-Object { Write-Host "[FRONTEND] $_" } }
        Write-Host "Removing frontend job: $($FeJob.Name)..."; Remove-Job -Job $FeJob -Force -ErrorAction SilentlyContinue; Write-Host "Frontend job '$($FeJob.Name)' removed." -ForegroundColor Green
    }

    # --- Backend Job Cleanup (Port 8091) ---
    if ($BeJob) {
        Write-Host "Processing backend job: $($BeJob.Name) (ID: $($BeJob.Id), Initial State: $($BeJob.State))"
        if ($BeJob.State -in ('Running', 'NotStarted', 'Blocked', 'Suspended', 'Suspending')) {
            Write-Host "Attempting to stop backend job: $($BeJob.Name)..."
            Stop-Job -Job $BeJob -Force -ErrorAction SilentlyContinue
            Write-Host "Backend job '$($BeJob.Name)' stop signal sent. Waiting..."
            Start-Sleep -Seconds 3
        }

        Write-Host "Checking port 8091 for lingering backend processes..."
        $maxAttempts = 3
        for ($i = 1; $i -le $maxAttempts; $i++) {
            $processesOnPort8091 = Get-NetTCPConnection -LocalPort 8091 -State Listen -ErrorAction SilentlyContinue |
                                   Select-Object -ExpandProperty OwningProcess | Get-Unique
            if ($processesOnPort8091.Count -eq 0) { Write-Host "Port 8091 is clear." -ForegroundColor Green; break }
            Write-Warning "Attempt $i/${maxAttempts}: Backend process(es) still on port 8091 (PIDs: $($processesOnPort8091 -join ', ')). Terminating..."
            foreach ($pidToKill in $processesOnPort8091) { Write-Host "Killing PID $pidToKill on port 8091..."; Stop-Process -Id $pidToKill -Force -ErrorAction SilentlyContinue }
            if ($i -lt $maxAttempts) { Start-Sleep -Seconds 2 }
            elseif ($processesOnPort8091.Count -gt 0) { Write-Error "Failed to clear port 8091 for backend after $maxAttempts attempts." }
        }

        if ($BeJob.HasMoreData) { Write-Host "Final output from $($BeJob.Name):"; Receive-Job -Job $BeJob | ForEach-Object { Write-Host "[BACKEND] $_" } }
        Write-Host "Removing backend job: $($BeJob.Name)..."; Remove-Job -Job $BeJob -Force -ErrorAction SilentlyContinue; Write-Host "Backend job '$($BeJob.Name)' removed." -ForegroundColor Green
    }

    Write-Host "Cleanup attempt complete. Check task manager or netstat if issues persist." -ForegroundColor Yellow
    Write-Host "Script finished."
}

try {
    # Trap Ctrl+C (SIGINT). This exception is thrown when Ctrl+C is pressed.
    trap [System.Management.Automation.PipelineStoppedException] {
        Write-Warning "`nCtrl+C detected! Script will attempt to clean up and exit."
        $scriptInterrupted = $true # Signal the main loop to exit. Cleanup will happen in 'finally'.
        # Do NOT call Stop-And-Remove-Jobs or exit from here.
    }

    # --- Start Frontend Server as a Background Job ---
    try {
        Write-Host "Starting Vue.js frontend development server (npm run dev) as a job from '$scriptRootPath'..."
        $frontendJob = Start-Job -Name "FrontendDevServer" -ScriptBlock {
            param($path)
            Set-Location -Path $path
            npm run dev *>&1
        } -ArgumentList $scriptRootPath

        if ($frontendJob) {
            Write-Host "Frontend server job started (Name: $($frontendJob.Name), ID: $($frontendJob.Id)). Output will stream below." -ForegroundColor Green
            Write-Host "Access frontend at http://localhost:3000 (usually)" -ForegroundColor Cyan
        } else { Write-Warning "Failed to start frontend server job." }
    } catch { Write-Error "Exception while starting frontend server job: $($_.Exception.Message)"; $frontendJob = $null }

    Write-Host ""

    # --- Start Backend Server as a Background Job ---
    $goApiDir = Join-Path $scriptRootPath "src/api"
    if (-not (Test-Path $goApiDir -PathType Container)) {
        Write-Error "Error: Go API directory not found at '$goApiDir'. Cannot start backend."; $backendJob = $null
    } else {
        try {
            Write-Host "Starting Go backend API server (go run . in '$goApiDir') as a job..."
            $backendJob = Start-Job -Name "BackendDevServer" -ScriptBlock {
                param($path)
                Set-Location -Path $path
                go run . *>&1
            } -ArgumentList $goApiDir

            if ($backendJob) {
                Write-Host "Backend server job started (Name: $($backendJob.Name), ID: $($backendJob.Id)). Output will stream below." -ForegroundColor Green
                Write-Host "Go API should be available at http://localhost:8091 (usually proxied by Vite)" -ForegroundColor Cyan
            } else { Write-Warning "Failed to start backend server job." }
        } catch { Write-Error "Exception while starting backend server job: $($_.Exception.Message)"; $backendJob = $null }
    }

    Write-Host ""; Write-Host "------------------------------------------------------------"

    if ($frontendJob -or $backendJob) {
        Write-Host "Servers are running as background jobs. Output is being streamed." -ForegroundColor Cyan
        Write-Host "Press 's' (and then Enter) to stop all servers and exit." -ForegroundColor Cyan
        Write-Host "Press Ctrl+C to interrupt and stop all servers." -ForegroundColor Cyan
        Write-Host "If output stops, a server might have errored or finished." -ForegroundColor Yellow; Write-Host ""

        while (-not $scriptInterrupted) { # Loop until Ctrl+C or 's' sets the flag
            $outputReceivedThisCycle = $false
            if ($frontendJob -and $frontendJob.HasMoreData) { Receive-Job -Job $frontendJob -Keep | ForEach-Object { Write-Host "[FRONTEND] $_"; $outputReceivedThisCycle = $true } }
            if ($backendJob -and $backendJob.HasMoreData) { Receive-Job -Job $backendJob -Keep | ForEach-Object { Write-Host "[BACKEND] $_"; $outputReceivedThisCycle = $true } }

            if ($Host.UI.RawUI.KeyAvailable) {
                $key = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
                if ($key.Character -eq 's') {
                    Write-Host ""; $confirmInput = Read-Host -Prompt "You pressed 's'. Enter to confirm stop, any other key + Enter to continue"
                    if ($confirmInput -eq "") { Write-Host "`nStop command confirmed." -ForegroundColor Yellow; $scriptInterrupted = $true; break }
                    else { Write-Host "Stop command cancelled." -ForegroundColor Yellow }
                }
            }

            $frontendJobDone = $true; if ($frontendJob -and $frontendJob.State -in ('Running', 'NotStarted', 'Blocked')) { $frontendJobDone = $false }
            $backendJobDone = $true;  if ($backendJob -and $backendJob.State -in ('Running', 'NotStarted', 'Blocked')) { $backendJobDone = $false }

            if (($frontendJob -eq $null -or $frontendJobDone) -and ($backendJob -eq $null -or $backendJobDone)) {
                $allJobsEffectivelyCompleted = $true
                if($frontendJob -ne $null -and -not $frontendJobDone) {$allJobsEffectivelyCompleted = $false}
                if($backendJob -ne $null -and -not $backendJobDone) {$allJobsEffectivelyCompleted = $false}
                if($allJobsEffectivelyCompleted) { Write-Host "`nServer jobs appear to have completed or failed." -ForegroundColor Green; $scriptInterrupted = $true; break }
            }

            if (-not $outputReceivedThisCycle) { Start-Sleep -Milliseconds 200 }
        }
    } else { Write-Warning "Neither server job could be started. Please check errors." }

} finally {
    # This 'finally' block will execute when the 'try' block is exited,
    # either because the 'while' loop broke (due to 's' or jobs completing)
    # or because an exception occurred (like Ctrl+C, which sets $scriptInterrupted via the trap).
    Stop-And-Remove-Jobs -FeJob $frontendJob -BeJob $backendJob
}
