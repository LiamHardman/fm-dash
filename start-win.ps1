# Set Execution Policy for this session
Set-ExecutionPolicy -Scope Process -ExecutionPolicy Bypass -Force

# Variables
$FacepackURL = "http://api.fm-dash.com/regen_facepack/regen.zip"
$FacepackDir = "fifa_card_assets\player_img"

# Confirm Installation Function
Function Confirm-Installation {
    Param ([string]$message)
    $response = Read-Host -Prompt "$message (yes/no)"
    return $response
}

# Install Facepack Function
Function Install-Facepack {
    $response = Confirm-Installation "Do you wish to proceed with the installation of the Facepack?"
    if ($response -eq 'yes') {
        # Ensure the directory exists
        New-Item -ItemType Directory -Force -Path $FacepackDir

        # Download and extract the facepack
        Write-Host "Downloading Facepack..."
        $ZipPath = Join-Path -Path $FacepackDir -ChildPath "regen.zip"
        Invoke-WebRequest -Uri $FacepackURL -OutFile $ZipPath

        Write-Host "Extracting Facepack..."
        Expand-Archive -Path $ZipPath -DestinationPath $FacepackDir -Force

        Write-Host "Facepack installation complete!"
    } elseif ($response -eq 'no') {
        Write-Host "Facepack installation aborted."
    } else {
        Write-Host "Please answer yes or no."
    }
}

# Check and Install Python and Pip, and Install Requirements Function
Function Install-Requirements {
    $pythonInstalled = $null -ne (Get-Command python -ErrorAction SilentlyContinue)
    $pipInstalled = $null -ne (Get-Command pip -ErrorAction SilentlyContinue)

    if (-not $pythonInstalled) {
        Write-Host "Python is not installed or not in the PATH. Please install Python from https://www.python.org/downloads/"
        exit
    }

    if (-not $pipInstalled) {
        Write-Host "pip is not installed."
        exit
    }

    if (Test-Path "requirements.txt") {
        Write-Host "Installing requirements from requirements.txt..."
        python -m pip install -r requirements.txt
    } else {
        Write-Host "requirements.txt not found. Skipping pip install."
    }
}

# Run the server (app.py) Function
Function Run-Server {
    if (Test-Path "app.py") {
        Write-Host "Running app.py..."
        python app.py
    } else {
        Write-Host "app.py not found. Skipping execution."
    }
}

# Main script logic with command-line arguments
switch ($args[0]) {
    "--install" {
        Install-Requirements
    }
    "--start" {
        Run-Server
    }
    "--facepack_install" {
        Install-Facepack
    }
    default {
        Write-Host "Usage: powershell .\run_win [OPTION]"
        Write-Host "--install: Check and install Python, pip, and project requirements."
        Write-Host "--start  : Start the server directly (assuming everything is installed)."
        Write-Host "--facepack_install: Download and install the facepack to the specified directory."
        Write-Host "Note: This script assumes Python and pip are installed and accessible from PATH."
    }
}

# Keep the window open until the user wants to exit
Read-Host -Prompt "Press Enter to exit"
