package core

import (
	"strings"
	"sync"
	"time"
)

type SessionManager struct {
	sessions map[string]*Session
	mu       sync.RWMutex
	onAdd    func(*Session)
	onUpdate func(*Session)
	onRemove func(string)
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*Session),
	}
}

func (sm *SessionManager) SetOnAdd(callback func(*Session)) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.onAdd = callback
}

func (sm *SessionManager) SetOnUpdate(callback func(*Session)) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.onUpdate = callback
}

func (sm *SessionManager) SetOnRemove(callback func(string)) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.onRemove = callback
}

func (sm *SessionManager) AddSession(session *Session) {
	sm.mu.Lock()
	sm.sessions[session.PreyID] = session
	callback := sm.onAdd
	sm.mu.Unlock()

	if callback != nil {
		callback(session)
	}
}

func (sm *SessionManager) GetSession(preyID string) (*Session, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	session, ok := sm.sessions[preyID]
	return session, ok
}

func (sm *SessionManager) GetAllSessions() []*Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	sessions := make([]*Session, 0, len(sm.sessions))
	for _, s := range sm.sessions {
		sessions = append(sessions, s)
	}
	return sessions
}

func (sm *SessionManager) UpdateSession(session *Session) {
	sm.mu.Lock()
	sm.sessions[session.PreyID] = session
	callback := sm.onUpdate
	sm.mu.Unlock()

	if callback != nil {
		callback(session)
	}
}

func (sm *SessionManager) MarkOffline(preyID string) {
	sm.mu.Lock()
	if session, ok := sm.sessions[preyID]; ok {
		session.Status = StatusOffline
		callback := sm.onUpdate
		sm.mu.Unlock()
		if callback != nil {
			callback(session)
		}
	} else {
		sm.mu.Unlock()
	}
}

func (sm *SessionManager) RemoveSession(preyID string) {
	sm.mu.Lock()
	delete(sm.sessions, preyID)
	callback := sm.onRemove
	sm.mu.Unlock()

	if callback != nil {
		callback(preyID)
	}
}

func (sm *SessionManager) CheckTimeout(timeout time.Duration) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	now := time.Now()
	for _, session := range sm.sessions {
		if session.Status == StatusOnline && now.Sub(session.LastSeen) > timeout {
			session.Status = StatusLost
			if sm.onUpdate != nil {
				go sm.onUpdate(session)
			}
		}
	}
}

func ParsePreyID(preyID string) (mac, ip, os string) {
	parts := strings.Split(preyID, " ")
	if len(parts) >= 1 {
		macip := parts[0]
		for i := 0; i < len(macip); i++ {
			if macip[i] >= '0' && macip[i] <= '9' {
				if i > 0 {
					mac = macip[:i]
					ip = macip[i:]
				}
				break
			}
		}
	}
	if len(parts) >= 2 {
		os = parts[1]
	} else {
		os = "Unknown"
	}
	return
}

