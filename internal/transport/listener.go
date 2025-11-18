package transport

import (
	"fmt"
	"sync"

	nkn "github.com/nknorg/nkn-sdk-go"
)

type Listener struct {
	tm       *TransportManager
	client   *nkn.MultiClient
	running  bool
	mu       sync.Mutex
	stopChan chan struct{}
}

func NewListener(tm *TransportManager) *Listener {
	return &Listener{
		tm:       tm,
		stopChan: make(chan struct{}),
	}
}

func (l *Listener) Start(identifier string, onMessage func(*nkn.Message)) error {
	l.mu.Lock()
	if l.running {
		l.mu.Unlock()
		return fmt.Errorf("listener already running")
	}

	client, err := l.tm.CreateClient(identifier)
	if err != nil {
		l.mu.Unlock()
		return err
	}

	l.client = client
	l.running = true
	l.mu.Unlock()

	<-client.OnConnect.C

	go func() {
		for {
			select {
			case <-l.stopChan:
				return
			case msg := <-client.OnMessage.C:
				if onMessage != nil {
					onMessage(msg)
				}
			}
		}
	}()

	return nil
}

func (l *Listener) Stop() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if !l.running {
		return fmt.Errorf("listener not running")
	}

	close(l.stopChan)
	if l.client != nil {
		l.client.Close()
	}
	l.running = false
	return nil
}

func (l *Listener) IsRunning() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.running
}

