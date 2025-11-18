package transport

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"math/rand"
	"time"

	nkn "github.com/nknorg/nkn-sdk-go"
)

type TransportManager struct {
	account      *nkn.Account
	clientConfig *nkn.ClientConfig
	listener     *nkn.MultiClient
	seedID       string
	threads      int
}

func NewTransportManager(seedID string, threads int) (*TransportManager, error) {
	seed, err := hex.DecodeString(seedID)
	if err != nil {
		return nil, fmt.Errorf("invalid seed: %w", err)
	}

	account, err := nkn.NewAccount(seed)
	if err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	clientConfig := &nkn.ClientConfig{
		SeedRPCServerAddr:       nil,
		RPCTimeout:              100000,
		RPCConcurrency:          5,
		MsgChanLen:              4096,
		ConnectRetries:          10,
		MsgCacheExpiration:      300000,
		MsgCacheCleanupInterval: 60000,
		WsHandshakeTimeout:      100000,
		WsWriteTimeout:          100000,
		MinReconnectInterval:    100,
		MaxReconnectInterval:    10000,
		MessageConfig:           nil,
		SessionConfig:           nil,
	}

	return &TransportManager{
		account:      account,
		clientConfig: clientConfig,
		seedID:       seedID,
		threads:      threads,
	}, nil
}

func (tm *TransportManager) CreateClient(identifier string) (*nkn.MultiClient, error) {
	client, err := nkn.NewMultiClient(tm.account, identifier, tm.threads, false, tm.clientConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}
	return client, nil
}

func (tm *TransportManager) CreateRandomClient() (*nkn.MultiClient, error) {
	randomID := generateRandomID()
	return tm.CreateClient(randomID)
}

func (tm *TransportManager) GetAccount() *nkn.Account {
	return tm.account
}

func (tm *TransportManager) GetConfig() *nkn.ClientConfig {
	return tm.clientConfig
}

func generateRandomID() string {
	str := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < 32; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	ctx := md5.New()
	ctx.Write(result)
	return hex.EncodeToString(ctx.Sum(nil))
}
