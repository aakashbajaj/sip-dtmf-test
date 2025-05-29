package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/emiago/sipgo"
	"github.com/emiago/sipgo/sip"
	"github.com/gordonklaus/portaudio"
)

const (
	// Audio configuration
	sampleRate = 44100
	channels   = 2

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
	// Get local IP
	localIP := getLocalIP()
	log.Printf("Starting SIP server on %s:%d", localIP, sipPort)

	// Listen on UDP
	addr := fmt.Sprintf("%s:%d", localIP, sipPort)
	err := s.server.ListenAndServe(context.Background(), "udp", addr)
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	// Register handlers
	s.server.OnInvite(s.handleInvite)
	s.server.OnBye(s.handleBye)
	s.server.OnRegister(s.handleRegister)

	log.Printf("SIP server started successfully on %s", addr)
	return nil
}

func (s *SIPServer) handleInvite(req *sip.Request, tx sip.ServerTransaction) {
	log.Printf("Received INVITE from %s", req.From())

	// Send 180 Ringing
	res := sip.NewResponseFromRequest(req, 180, "Ringing", nil)
	if err := tx.Respond(res); err != nil {
		log.Printf("Failed to send 180 Ringing: %v", err)
		return
	}

	// Accept the call with 200 OK
	res = sip.NewResponseFromRequest(req, 200, "OK", nil)

	// Add Contact header
	contactHeader := &sip.ContactHeader{
		Address: sip.Uri{
			Scheme: "sip",
			Host:   getLocalIP(),
			Port:   sipPort,
		},
	}
	res.AppendHeader(contactHeader)

	// Add SDP for audio
	sdp := generateSDP(getLocalIP())
	res.SetBody([]byte(sdp))
	res.AppendHeader(&sip.NewHeader{
		HeaderName: "Content-Type",
		Contents:   "application/sdp",
	})

	if err := tx.Respond(res); err != nil {
		log.Printf("Failed to send 200 OK: %v", err)
		return
	}

	log.Println("Call accepted, playing audio...")

	// Play audio in a goroutine
	go playAudio()
}

func (s *SIPServer) handleBye(req *sip.Request, tx sip.ServerTransaction) {
	log.Printf("Received BYE from %s", req.From())

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

func playAudio() {
	// Initialize PortAudio
	if err := portaudio.Initialize(); err != nil {
		log.Printf("Failed to initialize PortAudio: %v", err)
		return
	}
	defer portaudio.Terminate()

	// Create output stream
	out := make([]float32, 512*channels)
	stream, err := portaudio.OpenDefaultStream(0, channels, sampleRate, len(out)/channels, &out)
	if err != nil {
		log.Printf("Failed to open audio stream: %v", err)
		return
	}
	defer stream.Close()

	// Start the stream
	if err := stream.Start(); err != nil {
		log.Printf("Failed to start audio stream: %v", err)
		return
	}
	defer stream.Stop()

	log.Println("Playing test tone (440Hz)...")

	// Generate and play a test tone (440Hz - A4 note)
	frequency := 440.0
	phase := 0.0
	phaseIncrement := 2.0 * math.Pi * frequency / float64(sampleRate)

	// Play for 10 seconds
	duration := 10 * time.Second
	startTime := time.Now()

	for time.Since(startTime) < duration {
		// Generate sine wave
		for i := 0; i < len(out); i += channels {
			sample := float32(math.Sin(phase) * 0.3) // 0.3 is volume
			for c := 0; c < channels; c++ {
				out[i+c] = sample
			}
			phase += phaseIncrement
			if phase >= 2.0*math.Pi {
				phase -= 2.0 * math.Pi
			}
		}

		// Write to audio output
		if err := stream.Write(); err != nil {
			log.Printf("Failed to write audio: %v", err)
			return
		}
	}

	log.Println("Audio playback completed")
}

func generateSDP(localIP string) string {
	// Simple SDP for audio session
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
a=sendrecv
`, localIP, localIP)
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
	log.Println("Starting SIP Audio Server...")

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

	log.Printf("SIP server is running on %s:%d", getLocalIP(), sipPort)
	log.Println("Press Ctrl+C to stop...")

	<-sigChan
	log.Println("Shutting down...")
}
