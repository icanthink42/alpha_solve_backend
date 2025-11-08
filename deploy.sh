#!/bin/bash
set -e

# Configuration
APP_NAME="alpha-solve"
APP_USER="ec2-user"
APP_DIR="/opt/alpha-solve"
SERVICE_FILE="alpha-solve.service"

echo "Starting deployment of $APP_NAME..."

# Create application directory if it doesn't exist
if [ ! -d "$APP_DIR" ]; then
    echo "Creating application directory at $APP_DIR"
    mkdir -p $APP_DIR
    chown $APP_USER:$APP_USER $APP_DIR
fi

# Stop the service if it's running
if systemctl is-active --quiet $APP_NAME; then
    echo "Stopping $APP_NAME service..."
    systemctl stop $APP_NAME
fi

# Copy the new binary
echo "Installing new binary..."
cp alpha_solve_server $APP_DIR/
chmod +x $APP_DIR/alpha_solve_server
chown $APP_USER:$APP_USER $APP_DIR/alpha_solve_server

# Install systemd service file
echo "Installing systemd service..."
cp $SERVICE_FILE /etc/systemd/system/
systemctl daemon-reload

# Enable and start the service
echo "Starting $APP_NAME service..."
systemctl enable $APP_NAME
systemctl start $APP_NAME

# Check service status
echo "Checking service status..."
sleep 2
if systemctl is-active --quiet $APP_NAME; then
    echo "✓ Deployment successful! $APP_NAME is running."
    systemctl status $APP_NAME --no-pager
else
    echo "✗ Deployment failed! Service is not running."
    systemctl status $APP_NAME --no-pager
    exit 1
fi

echo "Deployment complete!"

