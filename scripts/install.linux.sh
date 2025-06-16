#!/usr/bin/env bash
set -euo pipefail

# Configuration
SERVICE_USER="ezhttp"
SERVICE_GROUP="ezhttp"
INSTALL_DIR="/opt/ezhttp"
CONFIG_DIR="/etc/ezhttp"
LOG_DIR="/var/log/ezhttp"
CERT_DIR="/etc/ezhttp/certs"
PUBLIC_DIR="/opt/ezhttp/public"

# Check if running as root
if [[ $EUID -ne 0 ]]; then
   echo "This script must be run as root"
   exit 1
fi

# Detect distribution
if [ -f /etc/debian_version ]; then
    echo "Detected Debian-based distribution"
elif [ -f /etc/redhat-release ]; then
    echo "Detected RHEL-based distribution"
else
    echo "Unsupported distribution. This script supports Debian-based and RHEL-based distributions only."
    exit 1
fi

# Create EZhttp user and group
echo "Creating user and group: ${SERVICE_USER}"
if ! getent group "${SERVICE_GROUP}" >/dev/null 2>&1; then
    groupadd --system "${SERVICE_GROUP}"
    echo "Created group: ${SERVICE_GROUP}"
else
    echo "Group ${SERVICE_GROUP} already exists"
fi

if ! id -u "${SERVICE_USER}" >/dev/null 2>&1; then
    useradd --system --shell /usr/sbin/nologin --home-dir /nonexistent --gid "${SERVICE_GROUP}" --comment "EZhttp Service User" "${SERVICE_USER}"
    echo "Created user: ${SERVICE_USER}"
else
    echo "User ${SERVICE_USER} already exists"
fi

# Create directories
echo "Creating directories..."
mkdir -p "${INSTALL_DIR}" "${CONFIG_DIR}" "${LOG_DIR}" "${CERT_DIR}" "${PUBLIC_DIR}"

# Set ownership and permissions
chown -R "${SERVICE_USER}:${SERVICE_GROUP}" "${LOG_DIR}"
chown -R root:root "${INSTALL_DIR}" "${CONFIG_DIR}"
chown -R "${SERVICE_USER}:${SERVICE_GROUP}" "${CERT_DIR}"
chmod 755 "${INSTALL_DIR}" "${CONFIG_DIR}" "${PUBLIC_DIR}"
chmod 700 "${CERT_DIR}"
chmod 755 "${LOG_DIR}"

# Copy binary
if [ -f "ezhttp" ]; then
    echo "Installing ezhttp binary..."
    cp ezhttp "${INSTALL_DIR}/"
    chmod 755 "${INSTALL_DIR}/ezhttp"
    chown root:root "${INSTALL_DIR}/ezhttp"
else
    echo "Error: ezhttp binary not found in current directory"
    echo "Please build the ezhttp binary first with: make build"
    exit 1
fi

# Copy systemd unit file
if [ -f "ezhttp.systemd.service" ]; then
    echo "Installing systemd service..."
    cp ezhttp.systemd.service /etc/systemd/system/ezhttp.service
    chmod 644 /etc/systemd/system/ezhttp.service
else
    echo "Error: systemd service file not found"
    exit 1
fi

# Generate self-signed certificate (for testing only)
if [ ! -f "${CERT_DIR}/localhost.crt" ] || [ ! -f "${CERT_DIR}/localhost.key" ]; then
    echo "Generating self-signed certificate..."
    openssl req -x509 \
        -newkey rsa:4096 \
        -keyout "${CERT_DIR}/localhost.key" \
        -out "${CERT_DIR}/localhost.crt" \
        -days 365 \
        -nodes \
        -subj "/CN=localhost" >/dev/null 2>&1
    chown "${SERVICE_USER}:${SERVICE_GROUP}" "${CERT_DIR}"/localhost.*
    chmod 600 "${CERT_DIR}/localhost.key"
    chmod 644 "${CERT_DIR}/localhost.crt"
    echo "Self-signed certificate generated"
else
    echo "TLS certificates already exist"
fi

# Create sample config file
if [ ! -f "${CONFIG_DIR}/config.json" ]; then
    echo "Creating sample configuration..."
    cat > "${CONFIG_DIR}/config.json" << 'EOF'
{
  "listen_addr": "0.0.0.0",
  "listen_port": "8080",
  "rate_limit": {
    "enabled": true,
    "requests_per_minute": 60,
    "burst_size": 10
  }
}
EOF
    chmod 644 "${CONFIG_DIR}/config.json"
fi

# Reload systemd
systemctl daemon-reload

echo "Installation complete!"
echo ""
echo "Next steps:"
echo "1. Edit configuration: ${CONFIG_DIR}/config.json"
echo "2. Add your web files to: ${PUBLIC_DIR}/"
echo "3. Start the service: systemctl start ezhttp"
echo "4. Enable auto-start: systemctl enable ezhttp"
echo "5. Check status: systemctl status ezhttp"
echo ""
echo "For proxy mode, set MODE=proxy in the systemd service file"
