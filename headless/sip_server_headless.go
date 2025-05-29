package headless

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/emiago/sipgo"
	"github.com/emiago/sipgo/sip"
)

const (
	// SIP server configuration
	sipPort = 5060
)

type SIPServer struct {
	server *sipgo.Server
	ua     *sipgo.UserAgent
}

func NewSIPServer() (*SIPServer, error) {
	// Create user agent
	ua, err := sipgo.NewUA()
	if err != nil {
		return nil, fmt.Errorf("failed to create user agent: %w", err)
	}

	// Create server
	server, err := sipgo.NewServer(ua)
	if err != nil {
		return nil, fmt.Errorf("failed to create server: %w", err)
	}

	return &SIPServer{
		server: server,
		ua:     ua,
	}, nil
}

func (s *SIPServer) Start() error {
	// For GCP, use 0.0.0.0 to listen on all interfaces
	listenIP := "0.0.0.0"
	publicIP := getPublicIP()

	log.Printf("Starting SIP server on %s:%d", listenIP, sipPort)
	log.Printf("Public IP: %s", publicIP)

	// Listen on UDP
	addr := fmt.Sprintf("%s:%d", listenIP, sipPort)
	err := s.server.ListenAndServe(context.Background(), "udp", addr)
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	// Register handlers
	s.server.OnInvite(s.handleInvite)
	s.server.OnBye(s.handleBye)
	s.server.OnRegister(s.handleRegister)
	s.server.OnAck(s.handleAck)
	s.server.OnCancel(s.handleCancel)

	log.Printf("SIP server started successfully on %s", addr)
	return nil
}

func (s *SIPServer) handleInvite(req *sip.Request, tx sip.ServerTransaction) {
	log.Printf("Received INVITE from %s", req.From())
	log.Printf("To: %s", req.To())
	log.Printf("Call-ID: %s", req.CallID())

	// Send 100 Trying
	res := sip.NewResponseFromRequest(req, 100, "Trying", nil)
	if err := tx.Respond(res); err != nil {
		log.Printf("Failed to send 100 Trying: %v", err)
		return
	}

	// Send 180 Ringing
	res = sip.NewResponseFromRequest(req, 180, "Ringing", nil)
	if err := tx.Respond(res); err != nil {
		log.Printf("Failed to send 180 Ringing: %v", err)
		return
	}

	// Simulate ringing for 2 seconds
	time.Sleep(2 * time.Second)

	// Accept the call with 200 OK
	res = sip.NewResponseFromRequest(req, 200, "OK", nil)

	// Add Contact header
	publicIP := getPublicIP()
	contactHeader := &sip.ContactHeader{
		Address: sip.Uri{
			Scheme: "sip",
			Host:   publicIP,
			Port:   sipPort,
		},
	}
	res.AppendHeader(contactHeader)

	// Add SDP for audio
	sdp := generateSDP(publicIP)
	res.SetBody([]byte(sdp))
	contentTypeHeaderSDP := sip.ContentTypeHeader("application/sdp")
	res.AppendHeader(&contentTypeHeaderSDP)

	if err := tx.Respond(res); err != nil {
		log.Printf("Failed to send 200 OK: %v", err)
		return
	}

	log.Println("Call accepted successfully")
	log.Println("In a real scenario, audio would be played here")
	log.Println("Simulating call duration of 10 seconds...")

	// Simulate call duration
	go func() {
		time.Sleep(10 * time.Second)
		log.Println("Call simulation completed")
	}()
}

func (s *SIPServer) handleBye(req *sip.Request, tx sip.ServerTransaction) {
	log.Printf("Received BYE from %s", req.From())
	log.Printf("Call-ID: %s", req.CallID())

	// Send 200 OK
	res := sip.NewResponseFromRequest(req, 200, "OK", nil)
	if err := tx.Respond(res); err != nil {
		log.Printf("Failed to send 200 OK for BYE: %v", err)
	}

	log.Println("Call ended")
}

func (s *SIPServer) handleRegister(req *sip.Request, tx sip.ServerTransaction) {
	log.Printf("Received REGISTER from %s", req.From())

	// Send 200 OK
	res := sip.NewResponseFromRequest(req, 200, "OK", nil)
	if err := tx.Respond(res); err != nil {
		log.Printf("Failed to send 200 OK for REGISTER: %v", err)
	}
}

func (s *SIPServer) handleAck(req *sip.Request, tx sip.ServerTransaction) {
	log.Printf("Received ACK from %s", req.From())
	log.Printf("Call-ID: %s", req.CallID())
	log.Println("Call is now established")
}

func (s *SIPServer) handleCancel(req *sip.Request, tx sip.ServerTransaction) {
	log.Printf("Received CANCEL from %s", req.From())
	log.Printf("Call-ID: %s", req.CallID())

	// Send 200 OK for CANCEL
	res := sip.NewResponseFromRequest(req, 200, "OK", nil)
	if err := tx.Respond(res); err != nil {
		log.Printf("Failed to send 200 OK for CANCEL: %v", err)
	}

	// Send 487 Request Terminated for the original INVITE
	res = sip.NewResponseFromRequest(req, 487, "Request Terminated", nil)
	if err := tx.Respond(res); err != nil {
		log.Printf("Failed to send 487 Request Terminated: %v", err)
	}

	log.Println("Call cancelled")
}

func generateSDP(publicIP string) string {
	// Simple SDP for audio session
	// Using common audio codecs supported by Twilio
	return fmt.Sprintf(`v=0
o=- 0 0 IN IP4 %s
s=SIP Call
c=IN IP4 %s
t=0 0
m=audio 20000 RTP/AVP 0 8 101
a=rtpmap:0 PCMU/8000
a=rtpmap:8 PCMA/8000
a=rtpmap:101 telephone-event/8000
a=fmtp:101 0-16
a=ptime:20
a=sendrecv
`, publicIP, publicIP)
}

func getPublicIP() string {
	// Try to get the public IP from environment variable first
	if ip := os.Getenv("PUBLIC_IP"); ip != "" {
		return ip
	}

	// Otherwise, try to detect it
	return getLocalIP()
}

func getLocalIP() string {
	// Try to get a non-loopback IP address
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}

	return "127.0.0.1"
}

func main() {
	log.Println("Starting SIP Server (Headless Mode)...")
	log.Println("This version is designed for cloud deployment without audio devices")

	// Log the public IP if set
	if publicIP := os.Getenv("PUBLIC_IP"); publicIP != "" {
		log.Printf("Using PUBLIC_IP from environment: %s", publicIP)
	}

	// Create SIP server
	server, err := NewSIPServer()
	if err != nil {
		log.Fatal(err)
	}

	// Start server
	if err := server.Start(); err != nil {
		log.Fatal(err)
	}

	// Wait for interrupt
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	log.Printf("SIP server is running on port %d", sipPort)
	log.Println("Press Ctrl+C to stop...")

	<-sigChan
	log.Println("Shutting down...")
}
