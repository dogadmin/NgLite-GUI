package core

import "time"

type SessionStatus int

const (
	StatusOnline SessionStatus = iota
	StatusOffline
	StatusLost
)

func (s SessionStatus) String() string {
	switch s {
	case StatusOnline:
		return "在线"
	case StatusOffline:
		return "离线"
	case StatusLost:
		return "失联"
	default:
		return "未知"
	}
}

type Session struct {
	PreyID    string
	MAC       string
	IP        string
	OS        string
	Status    SessionStatus
	LastSeen  time.Time
	CreatedAt time.Time
}

