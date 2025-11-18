package proxy

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	nkn "github.com/nknorg/nkn-sdk-go"
)

const (
	Socks5Version = 0x05
	NoAuth        = 0x00
	ConnectCmd    = 0x01
	IPv4Address   = 0x01
	DomainName    = 0x03
	IPv6Address   = 0x04
)

type Socks5Server struct {
	listener   net.Listener
	preyID     string
	nknClient  *nkn.MultiClient
	running    bool
	mu         sync.Mutex
	sessions   map[string]*ProxySession
	onLog      func(string)
	port       int
}

type ProxySession struct {
	ID         string
	LocalConn  net.Conn
	NKNSession interface{}
	Target     string
	CreatedAt  time.Time
}

func NewSocks5Server(port int, preyID string, client *nkn.MultiClient, onLog func(string)) *Socks5Server {
	return &Socks5Server{
		port:      port,
		preyID:    preyID,
		nknClient: client,
		sessions:  make(map[string]*ProxySession),
		onLog:     onLog,
	}
}

func (s *Socks5Server) Start() error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("服务器已在运行")
	}
	s.mu.Unlock()

	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", s.port))
	if err != nil {
		return fmt.Errorf("无法启动监听: %v", err)
	}

	s.mu.Lock()
	s.listener = listener
	s.running = true
	s.mu.Unlock()

	s.log(fmt.Sprintf("SOCKS5 服务器已启动在 127.0.0.1:%d", s.port))

	go s.acceptLoop()
	return nil
}

func (s *Socks5Server) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return fmt.Errorf("服务器未运行")
	}

	s.running = false
	if s.listener != nil {
		s.listener.Close()
	}

	for _, session := range s.sessions {
		if session.LocalConn != nil {
			session.LocalConn.Close()
		}
		if session.NKNSession != nil {
			if closer, ok := session.NKNSession.(io.Closer); ok {
				closer.Close()
			}
		}
	}
	s.sessions = make(map[string]*ProxySession)

	s.log("SOCKS5 服务器已停止")
	return nil
}

func (s *Socks5Server) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

func (s *Socks5Server) acceptLoop() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			s.mu.Lock()
			running := s.running
			s.mu.Unlock()
			if !running {
				return
			}
			s.log(fmt.Sprintf("接受连接错误: %v", err))
			continue
		}

		go s.handleConnection(conn)
	}
}

func (s *Socks5Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	if err := s.handleHandshake(conn); err != nil {
		s.log(fmt.Sprintf("握手失败: %v", err))
		return
	}

	target, err := s.handleRequest(conn)
	if err != nil {
		s.log(fmt.Sprintf("请求处理失败: %v", err))
		return
	}

	s.log(fmt.Sprintf("正在连接到目标: %s", target))

	preyAddr := s.preyID + "." + s.nknClient.Address()
	
	dialConfig := &nkn.DialConfig{
		DialTimeout: 30000,
	}
	
	session, err := s.nknClient.DialWithConfig(preyAddr, dialConfig)
	if err != nil {
		s.log(fmt.Sprintf("NKN 连接失败: %v", err))
		s.sendConnectReply(conn, 0x04)
		return
	}
	defer session.Close()

	connectCmd := fmt.Sprintf("PROXY_CONNECT:%s", target)
	if _, err := session.Write([]byte(connectCmd)); err != nil {
		s.log(fmt.Sprintf("发送连接命令失败: %v", err))
		s.sendConnectReply(conn, 0x04)
		return
	}

	response := make([]byte, 1024)
	n, err := session.Read(response)
	if err != nil || n == 0 || string(response[:n]) != "OK" {
		s.log(fmt.Sprintf("连接目标失败: %s", string(response[:n])))
		s.sendConnectReply(conn, 0x04)
		return
	}

	if err := s.sendConnectReply(conn, 0x00); err != nil {
		s.log(fmt.Sprintf("发送响应失败: %v", err))
		return
	}

	s.log(fmt.Sprintf("代理隧道已建立: %s", target))

	s.forwardTraffic(conn, session, target)
}

func (s *Socks5Server) handleHandshake(conn net.Conn) error {
	buf := make([]byte, 257)
	
	n, err := conn.Read(buf)
	if err != nil {
		return err
	}

	if n < 2 || buf[0] != Socks5Version {
		return fmt.Errorf("无效的 SOCKS 版本")
	}

	_, err = conn.Write([]byte{Socks5Version, NoAuth})
	return err
}

func (s *Socks5Server) handleRequest(conn net.Conn) (string, error) {
	buf := make([]byte, 512)
	n, err := conn.Read(buf)
	if err != nil {
		return "", err
	}

	if n < 7 {
		return "", fmt.Errorf("请求太短")
	}

	if buf[0] != Socks5Version {
		return "", fmt.Errorf("无效的版本")
	}

	if buf[1] != ConnectCmd {
		return "", fmt.Errorf("仅支持 CONNECT 命令")
	}

	var target string
	var port uint16

	switch buf[3] {
	case IPv4Address:
		if n < 10 {
			return "", fmt.Errorf("IPv4 地址不完整")
		}
		target = fmt.Sprintf("%d.%d.%d.%d", buf[4], buf[5], buf[6], buf[7])
		port = binary.BigEndian.Uint16(buf[8:10])

	case DomainName:
		domainLen := int(buf[4])
		if n < 5+domainLen+2 {
			return "", fmt.Errorf("域名不完整")
		}
		target = string(buf[5 : 5+domainLen])
		port = binary.BigEndian.Uint16(buf[5+domainLen : 7+domainLen])

	case IPv6Address:
		return "", fmt.Errorf("暂不支持 IPv6")

	default:
		return "", fmt.Errorf("不支持的地址类型")
	}

	return fmt.Sprintf("%s:%d", target, port), nil
}

func (s *Socks5Server) sendConnectReply(conn net.Conn, status byte) error {
	reply := []byte{
		Socks5Version,
		status,
		0x00,
		IPv4Address,
		0, 0, 0, 0,
		0, 0,
	}
	_, err := conn.Write(reply)
	return err
}

func (s *Socks5Server) forwardTraffic(localConn net.Conn, nknSession interface{}, target string) {
	var wg sync.WaitGroup
	wg.Add(2)

	session := nknSession.(io.ReadWriteCloser)

	go func() {
		defer wg.Done()
		buf := make([]byte, 32768)
		for {
			n, err := localConn.Read(buf)
			if err != nil {
				if err != io.EOF {
					s.log(fmt.Sprintf("本地读取错误: %v", err))
				}
				session.Close()
				return
			}

			_, err = session.Write(buf[:n])
			if err != nil {
				s.log(fmt.Sprintf("NKN 写入错误: %v", err))
				return
			}
		}
	}()

	go func() {
		defer wg.Done()
		buf := make([]byte, 32768)
		for {
			n, err := session.Read(buf)
			if err != nil {
				if err != io.EOF {
					s.log(fmt.Sprintf("NKN 读取错误: %v", err))
				}
				localConn.Close()
				return
			}

			_, err = localConn.Write(buf[:n])
			if err != nil {
				s.log(fmt.Sprintf("本地写入错误: %v", err))
				return
			}
		}
	}()

	wg.Wait()
	s.log(fmt.Sprintf("连接已关闭: %s", target))
}

func (s *Socks5Server) log(msg string) {
	if s.onLog != nil {
		s.onLog(msg)
	}
}

func (s *Socks5Server) GetStats() (int, int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.sessions), s.port
}

