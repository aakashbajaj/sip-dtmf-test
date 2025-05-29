#!/bin/bash

# GCP Environment Setup Helper
# Source this file: source setup_gcp_env.sh

echo "GCP SIP Test Environment Setup"
echo "=============================="
echo ""

# Function to prompt for value with default
prompt_with_default() {
    local prompt=$1
    local default=$2
    local var_name=$3
    
    read -p "$prompt [$default]: " value
    value=${value:-$default}
    export $var_name="$value"
    echo "Set $var_name=$value"
}

# Check if already configured
if [ ! -z "$GCP_VM_IP" ] && [ ! -z "$TWILIO_ACCOUNT_SID" ]; then
    echo "Environment appears to be configured already:"
    echo "GCP_VM_IP: $GCP_VM_IP"
    echo "SIP_SERVER_URL: $SIP_SERVER_URL"
    echo ""
    read -p "Reconfigure? (y/N): " reconfigure
    if [ "$reconfigure" != "y" ] && [ "$reconfigure" != "Y" ]; then
        return 0 2>/dev/null || exit 0
    fi
fi

echo ""
echo "Step 1: GCP Configuration"
echo "-------------------------"

# GCP VM Configuration
prompt_with_default "Enter your GCP VM external IP" "your-gcp-vm-ip" "GCP_VM_IP"
prompt_with_default "Enter your GCP VM username" "$(whoami)" "GCP_VM_USER"

# Set SIP server URL based on GCP IP
export SIP_SERVER_URL="sip:test@$GCP_VM_IP:5060"
echo "Set SIP_SERVER_URL=$SIP_SERVER_URL"

echo ""
echo "Step 2: Twilio Configuration"
echo "----------------------------"

# Twilio Configuration
prompt_with_default "Enter your Twilio Account SID" "your-account-sid" "TWILIO_ACCOUNT_SID"
prompt_with_default "Enter your Twilio Auth Token" "your-auth-token" "TWILIO_AUTH_TOKEN"
prompt_with_default "Enter your Twilio phone number" "+1234567890" "TWILIO_PHONE_NUMBER"
prompt_with_default "Enter the phone number to call" "+1987654321" "TO_PHONE_NUMBER"

echo ""
echo "Step 3: Local Webhook Configuration"
echo "-----------------------------------"
echo "You need to run ngrok in another terminal: ngrok http 5000"
echo "Then enter the HTTPS URL provided by ngrok."
echo ""
prompt_with_default "Enter your ngrok webhook URL" "https://your-subdomain.ngrok.io/twiml" "WEBHOOK_URL"

echo ""
echo "Environment Configuration Complete!"
echo "==================================="
echo ""
echo "Summary:"
echo "--------"
echo "GCP_VM_IP: $GCP_VM_IP"
echo "GCP_VM_USER: $GCP_VM_USER"
echo "SIP_SERVER_URL: $SIP_SERVER_URL"
echo "TWILIO_PHONE_NUMBER: $TWILIO_PHONE_NUMBER"
echo "TO_PHONE_NUMBER: $TO_PHONE_NUMBER"
echo "WEBHOOK_URL: $WEBHOOK_URL"
echo ""
echo "Next Steps:"
echo "-----------"
echo "1. Deploy SIP server: ./deploy_to_gcp.sh"
echo "2. Start ngrok: ngrok http 5000"
echo "3. Run test: python3 twilio_call.py"
echo ""
echo "To save this configuration for next time, add to your .bashrc or .zshrc:"
echo "export GCP_VM_IP=\"$GCP_VM_IP\""
echo "export GCP_VM_USER=\"$GCP_VM_USER\""
echo "export SIP_SERVER_URL=\"$SIP_SERVER_URL\"" 