# SIP-DTMF Test System - GCP Deployment Guide

This guide explains how to deploy the SIP server on a GCP VM and run the Python client locally.

## Architecture Overview

```
┌─────────────────┐         ┌──────────────────┐         ┌─────────────────┐
│   Local Machine │         │      Twilio      │         │    GCP VM       │
│                 │         │                  │         │                 │
│  Python Script  │────────▶│  Makes Call     │────────▶│  SIP Server     │
│  + Flask        │         │  Using TwiML    │   SIP   │  (Headless)     │
│  + ngrok        │◀────────│                 │         │  Port 5060      │
└─────────────────┘  TwiML  └──────────────────┘         └─────────────────┘
```

## Prerequisites

### GCP VM Requirements
- Ubuntu 20.04 or later
- Static external IP address
- Firewall rules configured:
  - **UDP port 5060** (SIP signaling)
  - **UDP ports 20000-30000** (RTP media - optional for this test)
- SSH access configured

### Local Machine Requirements
- Python 3.7+
- ngrok installed
- Go 1.21+ (for building the server)
- Twilio account with credentials

## Step 1: Prepare GCP VM

### 1.1 Create Firewall Rules (if not already done)

```bash
# Allow SIP traffic
gcloud compute firewall-rules create allow-sip \
    --allow udp:5060 \
    --source-ranges 0.0.0.0/0 \
    --description "Allow SIP traffic"

# Allow RTP traffic (optional for this test)
gcloud compute firewall-rules create allow-rtp \
    --allow udp:20000-30000 \
    --source-ranges 0.0.0.0/0 \
    --description "Allow RTP traffic"
```

### 1.2 Note Your GCP VM Details

```bash
# Get your VM's external IP
gcloud compute instances list

# Note down:
# - External IP address
# - Username for SSH
```

## Step 2: Deploy SIP Server to GCP

### 2.1 Set Environment Variables

On your local machine:

```bash
export GCP_VM_IP="your-gcp-vm-external-ip"
export GCP_VM_USER="your-gcp-username"
```

### 2.2 Make Deployment Script Executable

```bash
chmod +x deploy_to_gcp.sh
```

### 2.3 Run Deployment

```bash
./deploy_to_gcp.sh
```

This script will:
- Build the headless SIP server for Linux
- Copy it to your GCP VM
- Set up a systemd service
- Start the SIP server

### 2.4 Verify Deployment

```bash
# Check if the service is running
ssh $GCP_VM_USER@$GCP_VM_IP "sudo systemctl status sip_server"

# View logs
ssh $GCP_VM_USER@$GCP_VM_IP "sudo journalctl -u sip_server -f"
```

## Step 3: Set Up Local Environment

### 3.1 Install Python Dependencies

```bash
pip install -r requirements.txt
```

### 3.2 Set Twilio Credentials

```bash
export TWILIO_ACCOUNT_SID="your_account_sid"
export TWILIO_AUTH_TOKEN="your_auth_token"
export TWILIO_PHONE_NUMBER="+1234567890"  # Your Twilio phone number
export TO_PHONE_NUMBER="+1987654321"      # Phone number to call
```

### 3.3 Set SIP Server URL

```bash
export SIP_SERVER_URL="sip:test@$GCP_VM_IP:5060"
```

### 3.4 Start ngrok

In a new terminal:

```bash
ngrok http 5000
```

Note the HTTPS URL provided by ngrok (e.g., `https://abc123.ngrok.io`)

### 3.5 Set Webhook URL

```bash
export WEBHOOK_URL="https://your-ngrok-subdomain.ngrok.io/twiml"
```

## Step 4: Run the Test

### 4.1 Start Python Script

```bash
python3 twilio_call.py
```

### 4.2 Make the Call

1. The script will start a Flask server on port 5000
2. Press Enter when prompted to initiate the call
3. The call flow will be:
   - Twilio calls the specified phone number
   - When answered, Twilio fetches TwiML from your ngrok URL
   - TwiML directs Twilio to connect to your GCP SIP server
   - SIP server accepts the call and simulates a 10-second call

### 4.3 Monitor the Call

Watch the logs in both terminals:
- Local Python script output
- GCP VM SIP server logs: `ssh $GCP_VM_USER@$GCP_VM_IP "sudo journalctl -u sip_server -f"`

## Troubleshooting

### SIP Server Not Reachable

1. **Check firewall rules:**
   ```bash
   gcloud compute firewall-rules list
   ```

2. **Test connectivity:**
   ```bash
   nc -u -v $GCP_VM_IP 5060
   ```

3. **Check if server is listening:**
   ```bash
   ssh $GCP_VM_USER@$GCP_VM_IP "sudo netstat -ulnp | grep 5060"
   ```

### Twilio Connection Issues

1. **Verify SIP URL format:**
   - Should be: `sip:test@your-external-ip:5060`
   - Not: `sip:test@localhost:5060`

2. **Check ngrok is running:**
   - Ensure the webhook URL is accessible
   - Test: `curl https://your-ngrok-url/twiml`

3. **Enable Twilio SIP debugging:**
   - Go to Twilio Console > Debugger
   - Check for SIP connection errors

### No Audio (Expected)

The headless SIP server doesn't play actual audio since GCP VMs don't have audio devices. It simulates the call flow and logs all SIP messages.

## Useful Commands

### Service Management

```bash
# Restart SIP server
ssh $GCP_VM_USER@$GCP_VM_IP "sudo systemctl restart sip_server"

# Stop SIP server
ssh $GCP_VM_USER@$GCP_VM_IP "sudo systemctl stop sip_server"

# Check service status
ssh $GCP_VM_USER@$GCP_VM_IP "sudo systemctl status sip_server"
```

### Log Viewing

```bash
# Follow logs in real-time
ssh $GCP_VM_USER@$GCP_VM_IP "sudo journalctl -u sip_server -f"

# View last 100 lines
ssh $GCP_VM_USER@$GCP_VM_IP "sudo journalctl -u sip_server -n 100"
```

### Manual Testing

You can manually test the SIP server using a SIP client:

```bash
# Using netcat to send OPTIONS request
echo -e "OPTIONS sip:test@$GCP_VM_IP:5060 SIP/2.0\r\nVia: SIP/2.0/UDP localhost:5061\r\nFrom: <sip:test@localhost>\r\nTo: <sip:test@$GCP_VM_IP>\r\nCall-ID: test123\r\nCSeq: 1 OPTIONS\r\nContent-Length: 0\r\n\r\n" | nc -u $GCP_VM_IP 5060
```

## Next Steps

1. **Add DTMF Support**: Implement DTMF tone detection in the SIP server
2. **Add Media Processing**: Use a media server for actual audio processing
3. **Implement Security**: Add SIP authentication and encryption
4. **Scale**: Deploy multiple SIP servers behind a load balancer

## Security Considerations

1. **Restrict Firewall Rules**: Instead of `0.0.0.0/0`, restrict to Twilio's IP ranges
2. **Use SIP Authentication**: Implement digest authentication
3. **Enable TLS**: Use SIP over TLS (SIPS) for encrypted signaling
4. **Monitor Access**: Set up logging and alerting for unauthorized access attempts 