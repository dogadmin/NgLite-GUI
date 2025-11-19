package socks5

import (
	"fmt"
	"sync"
)

// Manager manages SOCKS5 tunnels
type Manager struct {
	tunnels map[string]*Tunnel
	mu      sync.RWMutex
}

// NewManager creates a new SOCKS5 manager
func NewManager() *Manager {
	return &Manager{
		tunnels: make(map[string]*Tunnel),
	}
}

// StartTunnel starts a SOCKS5 tunnel for a prey
func (m *Manager) StartTunnel(preyID string, tunnel *Tunnel) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if _, exists := m.tunnels[preyID]; exists {
		return fmt.Errorf("tunnel already exists for prey %s", preyID)
	}
	
	m.tunnels[preyID] = tunnel
	
	go func() {
		if err := tunnel.Start(); err != nil {
			fmt.Printf("Tunnel error for prey %s: %v\n", preyID, err)
		}
	}()
	
	return nil
}

// StopTunnel stops a SOCKS5 tunnel
func (m *Manager) StopTunnel(preyID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	tunnel, exists := m.tunnels[preyID]
	if !exists {
		return fmt.Errorf("no tunnel found for prey %s", preyID)
	}
	
	err := tunnel.Close()
	delete(m.tunnels, preyID)
	
	return err
}

// GetTunnel returns a tunnel for a prey
func (m *Manager) GetTunnel(preyID string) (*Tunnel, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	tunnel, exists := m.tunnels[preyID]
	return tunnel, exists
}

// GetAllTunnels returns all active tunnels
func (m *Manager) GetAllTunnels() map[string]*Tunnel {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	tunnels := make(map[string]*Tunnel)
	for k, v := range m.tunnels {
		tunnels[k] = v
	}
	
	return tunnels
}

// StopAll stops all tunnels
func (m *Manager) StopAll() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	for preyID, tunnel := range m.tunnels {
		tunnel.Close()
		delete(m.tunnels, preyID)
	}
}

