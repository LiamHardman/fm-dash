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
    ip=$(prompt_for_ip)
    port=$(prompt_for_port)
    create_systemd_service_for_flask $ip $port
}
install_python() {
    if ! command -v python3.11 > /dev/null; then
        echo "Python 3.11 not found."
        confirm_install "Python 3.11"
        if command -v apt > /dev/null; then
            sudo apt update
            sudo apt install -y python3.11 python3.11-venv
        elif command -v yum > /dev/null; then
            sudo yum install -y python3.11
        elif command -v dnf > /dev/null; then
            sudo dnf install -y python3.11
        elif command -v pacman > /dev/null; then
            sudo pacman -Sy python
        else
            echo "Package manager not supported. Please install Python 3.11 manually."
            exit 1
        fi
    else
        echo "Python 3.11 is already installed."
    fi
}
install_dev_tools() {
    if ! command -v gcc > /dev/null; then
        echo "GCC not found."
        confirm_install "development tools and GCC"
        if command -v apt > /dev/null; then
            sudo apt install -y build-essential python3.11-dev gcc
        elif command -v yum > /dev/null; then
            sudo yum groupinstall -y 'Development Tools'
            sudo yum install -y python3.11-devel gcc
        elif command -v dnf > /dev/null; then
            sudo dnf groupinstall -y 'Development Tools'
            sudo dnf install -y python3.11-devel gcc
        elif command -v pacman > /dev/null; then
            sudo pacman -Sy base-devel gcc
        else
            echo "Package manager not supported. Please install development tools and GCC manually."
            exit 1
        fi
    else
        echo "GCC and development tools are already installed."
    fi
}

create_venv_and_install() {
    # Create a virtual environment if it doesn't exist
    if [ ! -d "venv" ]; then
        python3.11 -m venv venv
    fi

    # Activate the virtual environment
    if [ -f "venv/bin/activate" ]; then
        source venv/bin/activate

        # Check and install requirements
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

prompt_for_port() {
    read -p "Enter a port number for UWSGI or press enter to use the default (8980): " port
    if [ -z "$port" ]; then
        echo "8980"  # Default port
    else
        echo "$port"
    fi
}
install_facepack() {
    echo "Do you wish to proceed with the installation of the Facepack? (yes/no)"
    read yn
    case $yn in
        [Yy]* )
            # Define the directory and file paths
            DIRECTORY="fifa_card_assets/player_img"
            ZIPFILE="${DIRECTORY}/regen.zip"
            URL="http://api.fm-dash.com/regen_facepack/regen.zip"

            # Create the directory if it doesn't exist
            mkdir -p $DIRECTORY

            # Download the zip file
            echo "Downloading Facepack..."
            curl -o $ZIPFILE $URL || wget -O $ZIPFILE $URL

            # Extract the zip file
            echo "Extracting Facepack..."
            unzip -o $ZIPFILE -d $DIRECTORY

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
# Function to prompt for the IP address
prompt_for_ip() {
    read -p "Enter the IP address for Flask to run on (default 0.0.0.0): " ip
    if [ -z "$ip" ]; then
        echo "0.0.0.0"  # Default IP
    else
        echo "$ip"
    fi
}

# Function to prompt for the port
prompt_for_port() {
    read -p "Enter the port number for Flask to run on (default 8091): " port
    if [ -z "$port" ]; then
        echo "8091"
    else
        echo "$port"
    fi
}


run_server() {
    APP_DIR=$(pwd)
    if [ -f "$APP_DIR/app.py" ]; then
        echo "How do you want to run the app?"
        select server in "Flask (localhost)" "UWSGI"; do
            case $server in
                "Flask (localhost)")
                    echo "Running Flask Development Server..."
                    source $APP_DIR/venv/bin/activate
                    python $APP_DIR/app.py
                    break
                    ;;
                "UWSGI")
                    if [ -f "$APP_DIR/uwsgi.ini" ]; then
                        cd $APP_DIR
                        port=$(prompt_for_port)

                        # Define log file path
                        LOGFILE="/var/log/uwsgi/uwsgi.log"

                        # Ensure log directory exists
                        sudo mkdir -p $(dirname $LOGFILE)
                        sudo touch $LOGFILE

                        echo "Running app with UWSGI on port $port..."
                        source $APP_DIR/venv/bin/activate
                        uwsgi --ini ./uwsgi.ini --http-socket="0.0.0.0:$port"  --daemonize $LOGFILE > /dev/null 2>&1 &
                        echo $! > "$PIDFILE"
                        # Give user feedback
                        sleep 2
                        echo "uWSGI app running on http://0.0.0.0:$port"
                        echo "Logging to $LOGFILE"
                    else
                        echo "uwsgi.ini file not found. Please provide a uwsgi.ini configuration file in the current directory."
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


create_systemd_service_for_flask() {
    SERVICE_NAME="fm-dash"
    SERVICE_FILE="/etc/systemd/system/$SERVICE_NAME.service"
    USER="$(whoami)"
    WORKING_DIRECTORY="$(pwd)"
    PYTHON_EXEC="$VENV_PATH/bin/python"
    local ip="$1"
    local port="$2"
    local working_dir=$(pwd)
    local venv_path="$working_dir/venv"
    local app_path="$working_dir/app.py"
    local config_path="$working_dir/config.yml"

    echo "Creating systemd service file at $SERVICE_FILE"
    sudo bash -c "cat << EOF > $SERVICE_FILE
[Unit]
Description=FM Dash
After=network.target

[Service]
User=$USER
WorkingDirectory=$working_dir
ExecStart=$venv_path/bin/python $app_path
Environment="FLASK_ENV=production"
Environment="FLASK_RUN_HOST=$ip"
Environment="FLASK_RUN_PORT=$port"
Environment="FMD_CONF_LOCATION=$config_path"
Restart=always

[Install]
WantedBy=multi-user.target
EOF"

    echo "Reloading systemd daemon"
    sudo systemctl daemon-reload
    echo "Enabling $SERVICE_NAME service"
    sudo systemctl enable $SERVICE_NAME.service
    echo "$SERVICE_NAME service created and enabled"
}

start_service() {
    echo "Starting $SERVICE_NAME service..."
    sudo systemctl start $SERVICE_NAME.service
    echo "$SERVICE_NAME service started."
}

stop_service() {
    echo "Stopping $SERVICE_NAME service..."
    sudo systemctl stop $SERVICE_NAME.service
    echo "$SERVICE_NAME service stopped."
}

case "$1" in
    --install)
        install_all
        ;;
    --start)
        start_service
        ;;
    --stop)
        stop_service
        ;;
    --facepack_install)
        install_facepack
        ;;
    *)
        echo "Usage: $0 [OPTION]"
        echo "--install: Install Python, development tools, create a virtual environment, and create a systemd service for Flask."
        echo "--start  : Start the Flask systemd service."
        echo "--stop   : Stop the Flask systemd service."
        echo "--facepack_install: Install the facepack necessary for regen/backup images on the FIFA-style cards."
        ;;
esac
