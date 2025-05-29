#!/bin/bash

# GCP VM Deployment Script for SIP Server
# This script helps deploy the SIP server to your GCP VM

echo "GCP SIP Server Deployment"
echo "========================"
echo ""

# Configuration - Update these values
GCP_VM_IP="${GCP_VM_IP:-your-gcp-vm-ip}"
GCP_VM_USER="${GCP_VM_USER:-your-username}"
GCP_PROJECT_NAME="${GCP_PROJECT_NAME:-sip-dtmf-test}"

if [ "$GCP_VM_IP" == "your-gcp-vm-ip" ]; then
    echo "Error: Please set GCP_VM_IP environment variable"
    echo "export GCP_VM_IP=\"your.gcp.vm.ip\""
    exit 1
fi

echo "Deploying to: $GCP_VM_USER@$GCP_VM_IP"
echo ""

# Build the Go binary for Linux
echo "Building Go binary for Linux (headless version)..."
GOOS=linux GOARCH=amd64 go build -o sip_server_linux sip_server_headless.go
if [ $? -ne 0 ]; then
    echo "Failed to build Go binary"
    exit 1
fi

echo "Binary built successfully"

# Create deployment directory on GCP VM
echo "Creating deployment directory on GCP VM..."
ssh $GCP_VM_USER@$GCP_VM_IP "mkdir -p ~/$GCP_PROJECT_NAME"

# Copy the binary to GCP VM
echo "Copying binary to GCP VM..."
scp sip_server_linux $GCP_VM_USER@$GCP_VM_IP:~/$GCP_PROJECT_NAME/

# Create systemd service file locally
cat > sip_server.service << EOF
[Unit]
Description=SIP Audio Server (Headless)
After=network.target

[Service]
Type=simple
User=$GCP_VM_USER
WorkingDirectory=/home/$GCP_VM_USER/$GCP_PROJECT_NAME
Environment="PUBLIC_IP=$GCP_VM_IP"
ExecStart=/home/$GCP_VM_USER/$GCP_PROJECT_NAME/sip_server_linux
Restart=always
RestartSec=10
StandardOutput=syslog
StandardError=syslog
SyslogIdentifier=sip-server

[Install]
WantedBy=multi-user.target
EOF

# Copy service file to GCP VM
echo "Setting up systemd service..."
scp sip_server.service $GCP_VM_USER@$GCP_VM_IP:/tmp/

# Install and start the service
ssh $GCP_VM_USER@$GCP_VM_IP << 'ENDSSH'
# Make binary executable
chmod +x ~/sip-dtmf-test/sip_server_linux

# Install systemd service
sudo mv /tmp/sip_server.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable sip_server.service
sudo systemctl restart sip_server.service

# Check status
sudo systemctl status sip_server.service
ENDSSH

# Clean up local files
rm -f sip_server_linux sip_server.service

echo ""
echo "Deployment completed!"
echo ""
echo "To check logs on GCP VM:"
echo "ssh $GCP_VM_USER@$GCP_VM_IP \"sudo journalctl -u sip_server -f\""
echo ""
echo "To restart the service:"
echo "ssh $GCP_VM_USER@$GCP_VM_IP \"sudo systemctl restart sip_server\""
echo ""
echo "Update your local environment:"
echo "export SIP_SERVER_URL=\"sip:test@$GCP_VM_IP:5060\""
echo ""
echo "Make sure your GCP firewall rules allow:"
echo "- UDP port 5060 (SIP)"
echo "- UDP ports 20000-30000 (RTP media)" 