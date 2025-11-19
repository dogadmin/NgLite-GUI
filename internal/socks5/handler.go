package socks5

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"time"
)

// HandleSOCKS5WithTunnel handles SOCKS5 protocol and forwards through NKN tunnel
func HandleSOCKS5WithTunnel(clientConn net.Conn, nknSession net.Conn) error {
	defer clientConn.Close()
	defer nknSession.Close()
	
	log.Printf("[SOCKS5Handler] New SOCKS5 connection from %s", clientConn.RemoteAddr())
	
	// 1. Handle SOCKS5 authentication
	if err := handleSOCKS5Auth(clientConn); err != nil {
		log.Printf("[SOCKS5Handler] Auth failed: %v", err)
		return err
	}
	
	// 2. Handle SOCKS5 request and get target address
	target, err := handleSOCKS5Request(clientConn)
	if err != nil {
		log.Printf("[SOCKS5Handler] Request failed: %v", err)
		return err
	}
	
	log.Printf("[SOCKS5Handler] Target address: %s", target)
	
	// 3. Send target address to prey through NKN
	targetBytes := []byte(target)
	targetLen := uint16(len(targetBytes))
	
	// Send length (2 bytes) + target address
	if err := binary.Write(nknSession, binary.BigEndian, targetLen); err != nil {
		log.Printf("[SOCKS5Handler] Failed to send target length: %v", err)
		return err
	}
	
	if _, err := nknSession.Write(targetBytes); err != nil {
		log.Printf("[SOCKS5Handler] Failed to send target address: %v", err)
		return err
	}
	
	log.Printf("[SOCKS5Handler] Sent target address to prey: %s", target)
	
	// 4. Wait for prey confirmation (1 byte: 0=success, 1=failed)
	statusBuf := make([]byte, 1)
	if _, err := io.ReadFull(nknSession, statusBuf); err != nil {
		log.Printf("[SOCKS5Handler] Failed to read prey status: %v", err)
		sendSOCKS5Reply(clientConn, 0x01) // General failure
		return err
	}
	
	if statusBuf[0] != 0 {
		log.Printf("[SOCKS5Handler] Prey failed to connect to target")
		sendSOCKS5Reply(clientConn, 0x04) // Host unreachable
		return fmt.Errorf("prey connection failed")
	}
	
	log.Printf("[SOCKS5Handler] Prey connected to target successfully")
	
	// 5. Send SOCKS5 success reply to client
	if err := sendSOCKS5Reply(clientConn, 0x00); err != nil {
		log.Printf("[SOCKS5Handler] Failed to send success reply: %v", err)
		return err
	}
	
	log.Printf("[SOCKS5Handler] Starting bidirectional forwarding")
	
	// 6. Start bidirectional forwarding
	done := make(chan struct{}, 2)
	
	go func() {
		n, _ := io.Copy(nknSession, clientConn)
		log.Printf("[SOCKS5Handler] Client->NKN: %d bytes", n)
		done <- struct{}{}
	}()
	
	go func() {
		n, _ := io.Copy(clientConn, nknSession)
		log.Printf("[SOCKS5Handler] NKN->Client: %d bytes", n)
		done <- struct{}{}
	}()
	
	<-done
	log.Printf("[SOCKS5Handler] Connection closed")
	
	return nil
}

func handleSOCKS5Auth(conn net.Conn) error {
	buf := make([]byte, 257)
	
	// Read version and nmethods
	if _, err := io.ReadFull(conn, buf[:2]); err != nil {
		return err
	}
	
	version := buf[0]
	if version != 0x05 {
		return fmt.Errorf("unsupported SOCKS version: %d", version)
	}
	
	nMethods := int(buf[1])
	if _, err := io.ReadFull(conn, buf[:nMethods]); err != nil {
		return err
	}
	
	// Send no authentication required
	_, err := conn.Write([]byte{0x05, 0x00})
	return err
}

func handleSOCKS5Request(conn net.Conn) (string, error) {
	buf := make([]byte, 4)
	
	// Read request header
	if _, err := io.ReadFull(conn, buf); err != nil {
		return "", err
	}
	
	version := buf[0]
	if version != 0x05 {
		return "", fmt.Errorf("unsupported SOCKS version: %d", version)
	}
	
	cmd := buf[1]
	if cmd != 0x01 { // CONNECT
		sendSOCKS5Reply(conn, 0x07) // Command not supported
		return "", fmt.Errorf("unsupported command: %d", cmd)
	}
	
	// Parse address
	atyp := buf[3]
	var host string
	
	switch atyp {
	case 0x01: // IPv4
		addr := make([]byte, 4)
		if _, err := io.ReadFull(conn, addr); err != nil {
			return "", err
		}
		host = net.IP(addr).String()
		
	case 0x03: // Domain
		lenBuf := make([]byte, 1)
		if _, err := io.ReadFull(conn, lenBuf); err != nil {
			return "", err
		}
		addr := make([]byte, lenBuf[0])
		if _, err := io.ReadFull(conn, addr); err != nil {
			return "", err
		}
		host = string(addr)
		
	case 0x04: // IPv6
		addr := make([]byte, 16)
		if _, err := io.ReadFull(conn, addr); err != nil {
			return "", err
		}
		host = net.IP(addr).String()
		
	default:
		sendSOCKS5Reply(conn, 0x08) // Address type not supported
		return "", fmt.Errorf("unsupported address type: %d", atyp)
	}
	
	// Read port
	portBuf := make([]byte, 2)
	if _, err := io.ReadFull(conn, portBuf); err != nil {
		return "", err
	}
	port := binary.BigEndian.Uint16(portBuf)
	
	target := net.JoinHostPort(host, strconv.Itoa(int(port)))
	return target, nil
}

func sendSOCKS5Reply(conn net.Conn, rep byte) error {
	reply := []byte{
		0x05, // Version
		rep,  // Reply code
		0x00, // Reserved
		0x01, // IPv4
		0, 0, 0, 0, // Bind address
		0, 0, // Bind port
	}
	_, err := conn.Write(reply)
	return err
}

// HandlePreyConnection handles incoming connection on prey side
func HandlePreyConnection(nknSession net.Conn) error {
	defer nknSession.Close()
	
	log.Printf("[PreyHandler] New NKN session from %s", nknSession.RemoteAddr())
	
	// 1. Read target address length
	var targetLen uint16
	if err := binary.Read(nknSession, binary.BigEndian, &targetLen); err != nil {
		log.Printf("[PreyHandler] Failed to read target length: %v", err)
		return err
	}
	
	// 2. Read target address
	targetBytes := make([]byte, targetLen)
	if _, err := io.ReadFull(nknSession, targetBytes); err != nil {
		log.Printf("[PreyHandler] Failed to read target address: %v", err)
		return err
	}
	
	target := string(targetBytes)
	log.Printf("[PreyHandler] Target address: %s", target)
	
	// 3. Connect to real target
	targetConn, err := net.DialTimeout("tcp", target, 10*time.Second)
	if err != nil {
		log.Printf("[PreyHandler] Failed to connect to target %s: %v", target, err)
		// Send failure status
		nknSession.Write([]byte{1})
		return err
	}
	defer targetConn.Close()
	
	log.Printf("[PreyHandler] Connected to target %s", target)
	
	// 4. Send success status
	if _, err := nknSession.Write([]byte{0}); err != nil {
		log.Printf("[PreyHandler] Failed to send success status: %v", err)
		return err
	}
	
	log.Printf("[PreyHandler] Starting bidirectional forwarding")
	
	// 5. Start bidirectional forwarding
	done := make(chan struct{}, 2)
	
	go func() {
		n, _ := io.Copy(targetConn, nknSession)
		log.Printf("[PreyHandler] NKN->Target: %d bytes", n)
		done <- struct{}{}
	}()
	
	go func() {
		n, _ := io.Copy(nknSession, targetConn)
		log.Printf("[PreyHandler] Target->NKN: %d bytes", n)
		done <- struct{}{}
	}()
	
	<-done
	log.Printf("[PreyHandler] Connection to %s closed", target)
	
	return nil
}

