package connection

/*
Author: Bastien Faivre
Project: EPFL, DCL, Performance and Security Evaluation of Layer 2 Blockchain Systems
Date: September 2023
*/

import (
	"bufio"
	"bytes"
	"censorship-proxy/configuration"
	"censorship-proxy/logs"
	"encoding/hex"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
)

//------------------------------------------------------------------------------
// Private variables
//------------------------------------------------------------------------------

// logger used by the client connection handler
var clientLoggers *logs.Loggers

//------------------------------------------------------------------------------
// Private methods
//------------------------------------------------------------------------------

// proxyTargetToClient proxies data from the target to the client
func proxyTargetToClient(closeChannel chan bool, clientConn net.Conn, targetConn net.Conn) {
	io.Copy(clientConn, targetConn)
	clientLoggers.Info.Println("TargetToClient closed for client", clientConn.RemoteAddr())
	closeChannel <- true
}

// proxyClientToTarget proxies data from the client to the target if it is not censored
func proxyClientToTarget(closeChannel chan bool, clientConn net.Conn, targetConn net.Conn, configManager *configuration.ConfigManager) {
	clientReader := bufio.NewReader(clientConn)
	// check censorship on each message
	for {
		// read http request
		request, err := http.ReadRequest(clientReader)
		if err != nil {
			if err != io.EOF && !strings.Contains(err.Error(), "connection reset by peer") {
				clientLoggers.Error.Println("Error reading request:", err)
			}
			break
		}
		// get config
		config := configManager.GetConfig()
		censored := false
		if len(config.CensoredAddresses) > 0 {
			// read body
			body, err := io.ReadAll(request.Body)
			if err != nil {
				clientLoggers.Error.Println("Error reading body:", err)
				break
			}
			request.Body = io.NopCloser(bytes.NewBuffer(body))
			// unmarshal body
			var bodyMap map[string]interface{}
			err = json.Unmarshal(body, &bodyMap)
			if err != nil {
				clientLoggers.Error.Println("Error unmarshalling body:", err)
				break
			}
			// check censorship
			switch bodyMap["method"] {
			case "eth_sendTransaction":
				// TODO: implement censorship
			case "eth_sendRawTransaction":
				signedTxn := bodyMap["params"].([]interface{})[0]
				// remove the 0x prefix if present
				if signedTxn.(string)[:2] == "0x" {
					signedTxn = signedTxn.(string)[2:]
				}
				raw, err := hex.DecodeString(signedTxn.(string))
				if err != nil {
					clientLoggers.Error.Println("Error decoding signed transaction:", err)
					break
				}
				var tx *types.Transaction
				rlp.DecodeBytes(raw, &tx)
				signer := types.NewEIP155Signer(tx.ChainId())
				sender, err := signer.Sender(tx)
				if err != nil {
					clientLoggers.Error.Println("Error getting sender:", err)
					break
				}
				senderFormatted := strings.ToLower(sender.Hex())
				if senderFormatted[:2] == "0x" {
					senderFormatted = senderFormatted[2:]
				}
				// check if the sender is censored
				for _, address := range config.CensoredAddresses {
					addressFormatted := strings.ToLower(address)
					if addressFormatted[:2] == "0x" {
						addressFormatted = addressFormatted[2:]
					}
					if strings.EqualFold(senderFormatted, addressFormatted) {
						clientLoggers.Warning.Println("Censored transaction from", sender.Hex())
						censored = true
						break
					}
				}
			default:
			}
		}
		// write http request to target
		if !censored {
			err = request.Write(targetConn)
			if err != nil {
				clientLoggers.Error.Println("Error writing request:", err)
				break
			}
		}
	}
	clientLoggers.Info.Println("ClientToTarget closed for client", clientConn.RemoteAddr())
	closeChannel <- true
}

//------------------------------------------------------------------------------
// Public methods
//------------------------------------------------------------------------------

// InitClientLoggers initializes the loggers for the client connection handler
func InitClientLoggers(loggers *logs.Loggers) {
	clientLoggers = loggers
}

// HandleClientConnection handles a new client connection
func HandleClientConnection(conn net.Conn, configManager *configuration.ConfigManager) {
	defer conn.Close()
	// connect to the target
	targetConn, err := net.Dial("tcp", configManager.GetConfig().TargetAddress)
	if err != nil {
		clientLoggers.Error.Println("Error connecting to target:", err)
		return
	}
	// start goroutines to handle data transmission in both directions
	closeChannel := make(chan bool)
	go proxyTargetToClient(closeChannel, conn, targetConn)
	go proxyClientToTarget(closeChannel, conn, targetConn, configManager)
	// wait for the goroutines to finish
	<-closeChannel
	<-closeChannel
	clientLoggers.Info.Println("Connection of client", conn.RemoteAddr(), "closed")
}
