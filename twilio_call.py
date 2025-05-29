import os
from twilio.rest import Client
from twilio.twiml.voice_response import VoiceResponse, Dial, Sip
from flask import Flask, Response
import threading

# Twilio credentials - Set these as environment variables
TWILIO_ACCOUNT_SID = os.environ.get('TWILIO_ACCOUNT_SID')
TWILIO_AUTH_TOKEN = os.environ.get('TWILIO_AUTH_TOKEN')
TWILIO_PHONE_NUMBER = os.environ.get('TWILIO_PHONE_NUMBER')  # Your Twilio phone number

# Configuration
TO_PHONE_NUMBER = os.environ.get('TO_PHONE_NUMBER', '+1234567890')  # Phone number to call
SIP_SERVER_URL = os.environ.get('SIP_SERVER_URL', 'sip:test@localhost:5060')  # Your SIP server URL
WEBHOOK_URL = os.environ.get('WEBHOOK_URL', 'http://your-server.ngrok.io/twiml')  # Your webhook URL for TwiML

# Flask app for serving TwiML
app = Flask(__name__)

@app.route('/twiml', methods=['GET', 'POST'])
def generate_twiml():
    """Generate TwiML that connects to SIP server"""
    response = VoiceResponse()
    
    # Create a Dial verb
    dial = Dial()
    
    # Add SIP endpoint
    dial.sip(SIP_SERVER_URL)
    
    # Add the Dial to the response
    response.append(dial)
    
    return Response(str(response), mimetype='text/xml')

def make_outbound_call():
    """Make an outbound call using Twilio"""
    try:
        # Initialize Twilio client
        client = Client(TWILIO_ACCOUNT_SID, TWILIO_AUTH_TOKEN)
        
        # Make the call
        call = client.calls.create(
            to=TO_PHONE_NUMBER,
            from_=TWILIO_PHONE_NUMBER,
            url=WEBHOOK_URL,
            method='POST'
        )
        
        print(f"Call initiated! Call SID: {call.sid}")
        print(f"Status: {call.status}")
        
    except Exception as e:
        print(f"Error making call: {e}")

def run_webhook_server():
    """Run the Flask webhook server in a separate thread"""
    app.run(port=5000, debug=False)

if __name__ == "__main__":
    print("Twilio SIP Call Test")
    print("====================")
    print(f"SIP Server URL: {SIP_SERVER_URL}")
    print(f"Webhook URL: {WEBHOOK_URL}")
    print(f"Calling: {TO_PHONE_NUMBER}")
    print()
    
    # Check if credentials are set
    if not TWILIO_ACCOUNT_SID or not TWILIO_AUTH_TOKEN or not TWILIO_PHONE_NUMBER:
        print("Error: Please set TWILIO_ACCOUNT_SID, TWILIO_AUTH_TOKEN, and TWILIO_PHONE_NUMBER environment variables")
        exit(1)
    
    # Start webhook server in a separate thread
    webhook_thread = threading.Thread(target=run_webhook_server, daemon=True)
    webhook_thread.start()
    print("Webhook server started on http://localhost:5000/twiml")
    
    # Wait a moment for the server to start
    import time
    time.sleep(2)
    
    # Make the call
    input("Press Enter to make the outbound call...")
    make_outbound_call()
    
    # Keep the webhook server running
    input("Press Enter to stop the webhook server...") 