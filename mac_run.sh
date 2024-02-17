#!/bin/bash

PIDFILE="/tmp/uwsgi.pid"

confirm_install() {
    while true; do
        read -p "Do you wish to proceed with the installation of $1? (yes/no) " yn
        case $yn in
            [Yy]* ) return 0;;
            [Nn]* ) echo "Installation aborted."; exit;;
            * ) echo "Please answer yes or no.";;
        esac
    done
}

install_all() {
    install_python
    install_dev_tools
    create_venv_and_install
}

# Function to install Python 3.11
install_python() {
    if ! command -v python3.11 > /dev/null; then
        echo "Python 3.11 not found."
        confirm_install "Python 3.11"
        echo "Installing Python 3.11..."
        brew install python@3.11
    else
        echo "Python 3.11 is already installed."
    fi
}

# Function to install development tools (Xcode Command Line Tools)
install_dev_tools() {
    if ! command -v gcc > /dev/null; then
        echo "GCC not found."
        confirm_install "development tools"
        echo "Installing development tools..."
        xcode-select --install
    else
        echo "Development tools are already installed."
    fi
}

# Function to create a virtual environment and install packages
create_venv_and_install() {
    if [ ! -d "venv" ]; then
        python3.11 -m venv venv
    fi

    if [ -f "venv/bin/activate" ]; then
        source venv/bin/activate

        if [ -f "requirements.txt" ]; then
            echo "Checking and installing requirements from requirements.txt..."
            while IFS= read -r package || [[ -n "$package" ]]; do
                if ! pip freeze | grep -i "^${package%%=*}==" > /dev/null; then
                    echo "Installing $package..."
                    python -m pip install "$package"
                else
                    echo "$package already installed."
                fi
            done < requirements.txt
        else
            echo "requirements.txt not found. Skipping pip install."
        fi
    else
        echo "Failed to create virtual environment."
        exit 1
    fi
}

# Function to prompt for a port or use the default
prompt_for_port() {
    read -p "Enter a port number for UWSGI or press enter to use the default (8980): " port
    if [ -z "$port" ]; then
        echo "8980"  # Default port
    else
        echo "$port"
    fi
}

run_server() {
    if [ -f "app.py" ]; then
        echo "How do you want to run the app?"
        select server in "Flask (localhost)" "UWSGI"; do
            case $server in
                "Flask (localhost)")
                    echo "Running app with the built-in Flask server..."
                    python app.py
                    break
                    ;;
                "UWSGI")
                    if [ -f "uwsgi.ini" ]; then
                        port=$(prompt_for_port)
                        echo "Running app with UWSGI on port $port..."
                        uwsgi --ini uwsgi.ini --http-socket 0.0.0.0:$port > /dev/null 2>&1 &
                        echo $! > "$PIDFILE"
                        sleep 2
                        echo "uWSGI app running on http://0.0.0.0:$port"
                    else
                        echo "uwsgi.ini file not found. Please provide a uwsgi.ini configuration file."
                    fi
                    break
                    ;;
                *)
                    echo "Invalid option. Please choose a valid server."
                    ;;
            esac
        done
    else
        echo "app.py not found. Skipping execution."
    fi
}

# Stop Server Function
stop_server() {
    echo "Attempting to stop the UWSGI server..."
    if [ -f "$PIDFILE" ]; then
        PID=$(cat $PIDFILE)
        if ps -p $PID > /dev/null; then
            echo "Stopping UWSGI server with PID: $PID"
            kill -9 $PID
            while ps -p $PID > /dev/null 2>&1; do
                sleep 1
            done
            echo "UWSGI server stopped."
        else
            echo "No UWSGI server found running with PID: $PID"
        fi
        rm $PIDFILE
    else
        echo "No PID file found. Is UWSGI running?"
    fi
}

# Function to install the Facepack
install_facepack() {
    echo "Do you wish to proceed with the installation of the Facepack? (yes/no)"
    read yn
    case $yn in
        [Yy]* )
            DIRECTORY="fifa_card_assets/player_img"
            ZIPFILE="${DIRECTORY}/regen.zip"
            URL="http://api.fm-dash.com/regen_facepack/regen.zip"

            # Create the directory if it doesn't exist
            mkdir -p "$DIRECTORY"

            # Download the zip file
            echo "Downloading Facepack..."
            curl -o "$ZIPFILE" "$URL"

            # Extract the zip file
            echo "Extracting Facepack..."
            unzip -o "$ZIPFILE" -d "$DIRECTORY"

            echo "Facepack installation complete!"
            ;;

        [Nn]* )
            echo "Facepack installation aborted."
            ;;

        * )
            echo "Please answer yes or no."
            ;;
    esac
}

# Main Logic
case "$1" in 
    --install)
        install_all
        ;;
    --start)
        run_server
        ;;
    --stop)
        stop_server
        ;;
    --facepack_install)
        install_facepack
        ;;
    *)
        echo "Usage: $0 [OPTION]"
        echo "--install: Install Python, development tools, and create a virtual environment."
        echo "--start  : Start the server directly (skipping installation steps)."
        echo "--stop   : Stop the server if it is running."
        echo "--facepack_install: Download and install the facepack to the specified directory."
        ;;
esac
