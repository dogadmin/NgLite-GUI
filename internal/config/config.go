package config

import (
	"NGLite/conf"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"

	nkn "github.com/nknorg/nkn-sdk-go"
)

type Config struct {
	SeedID       string `json:"seed_id"`
	HunterID     string `json:"hunter_id"`
	AESKey       string `json:"aes_key"`
	TransThreads int    `json:"trans_threads"`
	RSAPublicKey string `json:"rsa_public_key"`
}

func DefaultConfig() *Config {
	return &Config{
		SeedID:       conf.Seedid,
		HunterID:     conf.Hunterid,
		AESKey:       conf.AesKey,
		TransThreads: conf.TransThreads,
		RSAPublicKey: conf.RsaPublicKey,
	}
}

func LoadConfig(filepath string) (*Config, error) {
	if filepath == "" {
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &cfg, nil
}

func (c *Config) Save(filepath string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

func GenerateNewSeed() (string, error) {
	account, err := nkn.NewAccount(nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate seed: %w", err)
	}
	return hex.EncodeToString(account.Seed()), nil
}
