#!/bin/bash

echo "SIP-DTMF Test System Helper"
echo "=========================="
echo ""

# Check if PortAudio is installed
if [[ "$OSTYPE" == "darwin"* ]]; then
    if ! brew list portaudio &>/dev/null; then
        echo "❌ PortAudio is not installed. Please run: brew install portaudio"
        exit 1
    else
        echo "✅ PortAudio is installed"
    fi
elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
    if ! ldconfig -p | grep -q portaudio; then
        echo "❌ PortAudio is not installed. Please run: sudo apt-get install portaudio19-dev"
        exit 1
    else
        echo "✅ PortAudio is installed"
    fi
fi

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.21 or later"
    exit 1
else
    echo "✅ Go is installed: $(go version)"
fi

# Check if Python is installed
if ! command -v python3 &> /dev/null; then
    echo "❌ Python 3 is not installed"
    exit 1
else
    echo "✅ Python is installed: $(python3 --version)"
fi

echo ""
echo "Environment Variables Check:"
echo "---------------------------"

# Check Twilio credentials
if [ -z "$TWILIO_ACCOUNT_SID" ]; then
    echo "❌ TWILIO_ACCOUNT_SID is not set"
else
    echo "✅ TWILIO_ACCOUNT_SID is set"
fi

if [ -z "$TWILIO_AUTH_TOKEN" ]; then
    echo "❌ TWILIO_AUTH_TOKEN is not set"
else
    echo "✅ TWILIO_AUTH_TOKEN is set"
fi

if [ -z "$TWILIO_PHONE_NUMBER" ]; then
    echo "❌ TWILIO_PHONE_NUMBER is not set"
else
    echo "✅ TWILIO_PHONE_NUMBER is set: $TWILIO_PHONE_NUMBER"
fi

if [ -z "$TO_PHONE_NUMBER" ]; then
    echo "❌ TO_PHONE_NUMBER is not set (using default)"
else
    echo "✅ TO_PHONE_NUMBER is set: $TO_PHONE_NUMBER"
fi

# Get local IP
LOCAL_IP=$(ifconfig | grep -Eo 'inet (addr:)?([0-9]*\.){3}[0-9]*' | grep -Eo '([0-9]*\.){3}[0-9]*' | grep -v '127.0.0.1' | head -n 1)
echo ""
echo "Local IP Address: $LOCAL_IP"

if [ -z "$SIP_SERVER_URL" ]; then
    echo "❌ SIP_SERVER_URL is not set"
    echo "   Suggested value: export SIP_SERVER_URL=\"sip:test@$LOCAL_IP:5060\""
else
    echo "✅ SIP_SERVER_URL is set: $SIP_SERVER_URL"
fi

if [ -z "$WEBHOOK_URL" ]; then
    echo "❌ WEBHOOK_URL is not set"
    echo "   You need to run ngrok: ngrok http 5000"
    echo "   Then: export WEBHOOK_URL=\"https://your-subdomain.ngrok.io/twiml\""
else
    echo "✅ WEBHOOK_URL is set: $WEBHOOK_URL"
fi

echo ""
echo "Quick Setup Commands:"
echo "--------------------"
echo "# Set up environment variables:"
echo "export TWILIO_ACCOUNT_SID=\"your_account_sid\""
echo "export TWILIO_AUTH_TOKEN=\"your_auth_token\""
echo "export TWILIO_PHONE_NUMBER=\"+1234567890\""
echo "export TO_PHONE_NUMBER=\"+1987654321\""
echo "export SIP_SERVER_URL=\"sip:test@$LOCAL_IP:5060\""
echo ""
echo "# In terminal 1 - Start ngrok:"
echo "ngrok http 5000"
echo ""
echo "# Then set webhook URL:"
echo "export WEBHOOK_URL=\"https://your-subdomain.ngrok.io/twiml\""
echo ""
echo "# In terminal 2 - Start SIP server:"
echo "go run sip_server.go"
echo ""
echo "# In terminal 3 - Run Python script:"
echo "python3 twilio_call.py" 