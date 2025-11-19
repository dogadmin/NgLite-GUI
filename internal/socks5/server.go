package socks5

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strconv"
)

const (
	socks5Version = 0x05
	
	// Auth methods
	authNoAuth       = 0x00
	authNoAcceptable = 0xFF
	
	// Commands
	cmdConnect      = 0x01
	cmdBind         = 0x02
	cmdUDPAssociate = 0x03
	
	// Address types
	atypIPv4   = 0x01
	atypDomain = 0x03
	atypIPv6   = 0x04
	
	// Reply codes
	repSuccess              = 0x00
	repGeneralFailure       = 0x01
	repConnectionNotAllowed = 0x02
	repNetworkUnreachable   = 0x03
	repHostUnreachable      = 0x04
	repConnectionRefused    = 0x05
	repTTLExpired           = 0x06
	repCommandNotSupported  = 0x07
	repAddressNotSupported  = 0x08
)

// Server represents a SOCKS5 server
type Server struct {
	listener net.Listener
	closed   bool
}

// NewServer creates a new SOCKS5 server
func NewServer(listenAddr string) (*Server, error) {
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %w", err)
	}
	
	return &Server{
		listener: listener,
		closed:   false,
	}, nil
}

// Serve starts accepting connections
func (s *Server) Serve() error {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.closed {
				return nil
			}
			return err
		}
		
		go s.handleConnection(conn)
	}
}

// Close closes the server
func (s *Server) Close() error {
	s.closed = true
	return s.listener.Close()
}

// Addr returns the listening address
func (s *Server) Addr() string {
	return s.listener.Addr().String()
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	
	// 1. Handle authentication
	if err := s.handleAuth(conn); err != nil {
		return
	}
	
	// 2. Handle request
	target, err := s.handleRequest(conn)
	if err != nil {
		return
	}
	
	// 3. Connect to target
	targetConn, err := net.Dial("tcp", target)
	if err != nil {
		s.sendReply(conn, repHostUnreachable)
		return
	}
	defer targetConn.Close()
	
	// 4. Send success reply
	if err := s.sendReply(conn, repSuccess); err != nil {
		return
	}
	
	// 5. Start forwarding
	s.forward(conn, targetConn)
}

func (s *Server) handleAuth(conn net.Conn) error {
	buf := make([]byte, 257)
	
	// Read auth request
	if _, err := io.ReadFull(conn, buf[:2]); err != nil {
		return err
	}
	
	version := buf[0]
	if version != socks5Version {
		return fmt.Errorf("unsupported SOCKS version: %d", version)
	}
	
	nMethods := int(buf[1])
	if _, err := io.ReadFull(conn, buf[:nMethods]); err != nil {
		return err
	}
	
	// Send auth response (no authentication required)
	_, err := conn.Write([]byte{socks5Version, authNoAuth})
	return err
}

func (s *Server) handleRequest(conn net.Conn) (string, error) {
	buf := make([]byte, 4)
	
	// Read request header
	if _, err := io.ReadFull(conn, buf); err != nil {
		return "", err
	}
	
	version := buf[0]
	if version != socks5Version {
		return "", fmt.Errorf("unsupported SOCKS version: %d", version)
	}
	
	cmd := buf[1]
	if cmd != cmdConnect {
		s.sendReply(conn, repCommandNotSupported)
		return "", fmt.Errorf("unsupported command: %d", cmd)
	}
	
	// Parse address
	atyp := buf[3]
	var host string
	var err error
	
	switch atyp {
	case atypIPv4:
		addr := make([]byte, 4)
		if _, err := io.ReadFull(conn, addr); err != nil {
			return "", err
		}
		host = net.IP(addr).String()
		
	case atypDomain:
		addrLen := make([]byte, 1)
		if _, err := io.ReadFull(conn, addrLen); err != nil {
			return "", err
		}
		addr := make([]byte, addrLen[0])
		if _, err := io.ReadFull(conn, addr); err != nil {
			return "", err
		}
		host = string(addr)
		
	case atypIPv6:
		addr := make([]byte, 16)
		if _, err := io.ReadFull(conn, addr); err != nil {
			return "", err
		}
		host = net.IP(addr).String()
		
	default:
		s.sendReply(conn, repAddressNotSupported)
		return "", fmt.Errorf("unsupported address type: %d", atyp)
	}
	
	// Read port
	portBuf := make([]byte, 2)
	if _, err = io.ReadFull(conn, portBuf); err != nil {
		return "", err
	}
	port := binary.BigEndian.Uint16(portBuf)
	
	target := net.JoinHostPort(host, strconv.Itoa(int(port)))
	return target, nil
}

func (s *Server) sendReply(conn net.Conn, rep byte) error {
	reply := []byte{
		socks5Version,
		rep,
		0x00, // Reserved
		atypIPv4,
		0, 0, 0, 0, // Bind address
		0, 0, // Bind port
	}
	_, err := conn.Write(reply)
	return err
}

func (s *Server) forward(client, target net.Conn) {
	done := make(chan struct{}, 2)
	
	go func() {
		io.Copy(target, client)
		done <- struct{}{}
	}()
	
	go func() {
		io.Copy(client, target)
		done <- struct{}{}
	}()
	
	<-done
}

