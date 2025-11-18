package proxy

import (
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
)

type TcpForwarder struct {
	sessions map[string]net.Conn
	mu       sync.RWMutex
}

func NewTcpForwarder() *TcpForwarder {
	return &TcpForwarder{
		sessions: make(map[string]net.Conn),
	}
}

func (f *TcpForwarder) HandleProxySession(session io.ReadWriteCloser) error {
	defer session.Close()

	buf := make([]byte, 1024)
	n, err := session.Read(buf)
	if err != nil {
		return fmt.Errorf("读取连接命令失败: %v", err)
	}

	cmd := string(buf[:n])
	if !strings.HasPrefix(cmd, "PROXY_CONNECT:") {
		return fmt.Errorf("无效的代理命令: %s", cmd)
	}

	target := strings.TrimPrefix(cmd, "PROXY_CONNECT:")
	
	targetConn, err := net.DialTimeout("tcp", target, 30*1e9)
	if err != nil {
		session.Write([]byte(fmt.Sprintf("FAILED:%v", err)))
		return fmt.Errorf("连接目标失败: %v", err)
	}
	defer targetConn.Close()

	if _, err := session.Write([]byte("OK")); err != nil {
		return fmt.Errorf("发送确认失败: %v", err)
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		buf := make([]byte, 32768)
		for {
			n, err := session.Read(buf)
			if err != nil {
				targetConn.Close()
				return
			}
			_, err = targetConn.Write(buf[:n])
			if err != nil {
				return
			}
		}
	}()

	go func() {
		defer wg.Done()
		buf := make([]byte, 32768)
		for {
			n, err := targetConn.Read(buf)
			if err != nil {
				session.Close()
				return
			}
			_, err = session.Write(buf[:n])
			if err != nil {
				return
			}
		}
	}()

	wg.Wait()
	return nil
}

func (f *TcpForwarder) Close() {
	f.mu.Lock()
	defer f.mu.Unlock()
	
	for _, conn := range f.sessions {
		conn.Close()
	}
	f.sessions = make(map[string]net.Conn)
}

