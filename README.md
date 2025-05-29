# SIP-DTMF Test System

This project demonstrates a system where Twilio makes an outbound call using TwiML that connects to a SIP server, which then plays audio through the local speaker.

## Components

1. **Python Script (`twilio_call.py`)**: Generates TwiML and initiates outbound calls via Twilio
2. **Go SIP Server (`sip_server.go`)**: Receives SIP calls and plays audio through local speakers

## Prerequisites

### For Python Script
- Python 3.7+
- Twilio account with:
  - Account SID
  - Auth Token
  - A Twilio phone number
- ngrok or similar service for exposing local webhook to internet

### For Go SIP Server
- Go 1.21+
- PortAudio library installed on your system

## Installation

### 1. Install PortAudio (required for Go SIP server)

**macOS:**
```bash
brew install portaudio
```

**Ubuntu/Debian:**
```bash
sudo apt-get install portaudio19-dev
```

**Windows:**
Download and install from [PortAudio website](http://www.portaudio.com/download.html)

### 2. Install Python dependencies
```bash
pip install -r requirements.txt
```

### 3. Install Go dependencies
```bash
go mod download
```

## Configuration

### 1. Set up environment variables for Twilio

```bash
export TWILIO_ACCOUNT_SID="your_account_sid"
export TWILIO_AUTH_TOKEN="your_auth_token"
export TWILIO_PHONE_NUMBER="+1234567890"  # Your Twilio phone number
export TO_PHONE_NUMBER="+1987654321"      # Phone number to call
```

### 2. Configure SIP server URL

The SIP server URL format should be: `sip:username@your-ip:5060`

For local testing:
```bash
export SIP_SERVER_URL="sip:test@192.168.1.100:5060"  # Replace with your local IP
```

### 3. Set up webhook URL

You need to expose your local webhook server to the internet. Using ngrok:

```bash
ngrok http 5000
```

Then set the webhook URL:
```bash
export WEBHOOK_URL="https://your-subdomain.ngrok.io/twiml"
```

## Usage

### 1. Start the Go SIP Server

```bash
go run sip_server.go
```

The server will:
- Start listening on port 5060 for SIP requests
- Display the local IP address it's listening on
- Wait for incoming SIP calls

### 2. Run the Python Script

In a new terminal:

```bash
python twilio_call.py
```

The script will:
- Start a Flask webhook server on port 5000
- Wait for you to press Enter to initiate the call
- Make an outbound call via Twilio
- Generate TwiML that connects to your SIP server

### 3. Call Flow

1. Twilio makes an outbound call to the specified phone number
2. When answered, Twilio fetches TwiML from your webhook
3. TwiML instructs Twilio to connect to your SIP server
4. SIP server accepts the call and plays a 440Hz test tone for 10 seconds

## Troubleshooting

### Common Issues

1. **"Failed to initialize PortAudio"**
   - Ensure PortAudio is properly installed
   - On macOS, you might need to grant microphone/audio permissions

2. **"Connection refused" on SIP server**
   - Check firewall settings
   - Ensure the SIP server IP in the environment variable matches your actual IP
   - Port 5060 might be blocked or in use

3. **Twilio webhook not reachable**
   - Ensure ngrok is running and the URL is correctly set
   - Check that the Flask server is running on port 5000

4. **No audio output**
   - Check system audio settings
   - Ensure no other application is using the audio device
   - Try adjusting the volume in the `playAudio()` function

## Customization

### Changing the Audio Output

In `sip_server.go`, modify the `playAudio()` function to:
- Play different frequencies
- Play WAV files
- Adjust volume
- Change duration

### Adding DTMF Support

To add DTMF tone detection/generation:
1. Implement DTMF tone generator in the Go server
2. Add SIP INFO or RFC 2833 support for DTMF
3. Handle DTMF events in the SIP message handlers

## Security Notes

- Never commit your Twilio credentials to version control
- Use environment variables or secure credential storage
- In production, use HTTPS for webhooks and consider SIP authentication
- Implement proper error handling and logging

## License

This is a test/demo project. Use at your own risk. 