#!/bin/bash
set -e

# Configuration
APP_NAME="alpha-solve"
APP_USER="ec2-user"
APP_DIR="/opt/alpha-solve"
SERVICE_FILE="alpha-solve.service"

echo "Starting deployment of $APP_NAME..."

# Install Atlas if not present
if ! command -v atlas &> /dev/null; then
    echo "Installing Atlas..."
    curl -sSf https://atlasgo.sh | sh
fi

# Install PostgreSQL client if not present
if ! command -v psql &> /dev/null; then
    echo "Installing PostgreSQL client..."
    yum install -y postgresql15
fi

# Create application directory if it doesn't exist
if [ ! -d "$APP_DIR" ]; then
    echo "Creating application directory at $APP_DIR"
    mkdir -p $APP_DIR
    chown $APP_USER:$APP_USER $APP_DIR
fi

# Check if .env file exists, if not create from example
if [ ! -f "$APP_DIR/.env" ]; then
    echo "WARNING: $APP_DIR/.env not found!"
    echo "Please create it with your database credentials:"
    echo "  sudo nano $APP_DIR/.env"
    echo "Example:"
    echo "  DATABASE_URL=postgres://postgres:PASSWORD@neelemanetdb.cwv4ig2icg26.us-east-1.rds.amazonaws.com:5432/alpha_solve?sslmode=require"
    echo ""
    echo "Deployment will continue, but the service may not start without valid credentials."
fi

# Stop the service if it's running
if systemctl is-active --quiet $APP_NAME; then
    echo "Stopping $APP_NAME service..."
    systemctl stop $APP_NAME
fi

# Copy the new binary and schema files
echo "Installing new binary and schema..."
cp alpha_solve_server $APP_DIR/
cp schema.sql $APP_DIR/
cp atlas.hcl $APP_DIR/
cp -r migrations $APP_DIR/ 2>/dev/null || echo "No migrations directory"
chmod +x $APP_DIR/alpha_solve_server
chown -R $APP_USER:$APP_USER $APP_DIR

# Run database migrations
echo "Running database setup and migrations..."
cd $APP_DIR
if [ -f "$APP_DIR/.env" ]; then
    source $APP_DIR/.env
    if [ -n "$DATABASE_URL" ]; then
        # Extract database name and connection details
        DB_NAME=$(echo $DATABASE_URL | sed -n 's|.*//[^/]*/\([^?]*\).*|\1|p')
        DB_HOST=$(echo $DATABASE_URL | sed -n 's|.*@\([^:/]*\).*|\1|p')
        DB_USER=$(echo $DATABASE_URL | sed -n 's|.*//\([^:@]*\).*|\1|p')
        POSTGRES_URL=$(echo $DATABASE_URL | sed "s|/$DB_NAME|/postgres|")

        echo "Creating database '$DB_NAME' if it doesn't exist..."
        psql "$POSTGRES_URL" -c "CREATE DATABASE $DB_NAME;" 2>/dev/null || echo "Database already exists or couldn't be created"

        echo "Running migrations..."
        atlas migrate apply --env prod --url "$DATABASE_URL" || echo "Warning: Migration failed, continuing anyway"
    else
        echo "DATABASE_URL not set in .env, skipping migrations"
    fi
else
    echo ".env file not found, skipping migrations"
fi
cd -

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

