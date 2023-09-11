package connection

/*
Author: Bastien Faivre
Project: EPFL, DCL, Performance and Security Evaluation of Layer 2 Blockchain Systems
Date: September 2023
*/

import (
	"censorship-proxy/configuration"
	"censorship-proxy/logs"
	"io"
	"net"
)

//------------------------------------------------------------------------------
// Constants
//------------------------------------------------------------------------------

// size of the buffer used to read data
const CLIENT_BUFFER_SIZE = 1048576 // 1MB

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
	closeChannel <- true
}

// proxyClientToTarget proxies data from the client to the target if it is not censored
func proxyClientToTarget(closeChannel chan bool, clientConn net.Conn, targetConn net.Conn, config configuration.Config) {
	// check censorship on each message
	for {
		// read data
		data := make([]byte, CLIENT_BUFFER_SIZE)
		n, err := clientConn.Read(data)
		if err != nil {
			clientLoggers.Error.Println("Error reading data:", err)
			closeChannel <- true
			return
		}
		// TODO: check censorship
		clientLoggers.Info.Println("Received data:", string(data[:n]))
		// write data
		_, err = targetConn.Write(data[:n])
		if err != nil {
			clientLoggers.Error.Println("Error writing data to target:", err)
			closeChannel <- true
			return
		}
	}
}

//------------------------------------------------------------------------------
// Public methods
//------------------------------------------------------------------------------

// InitClientLoggers initializes the loggers for the client connection handler
func InitClientLoggers(loggers *logs.Loggers) {
	clientLoggers = loggers
}

// HandleClientConnection handles a new client connection
func HandleClientConnection(conn net.Conn, config configuration.Config) {
	defer conn.Close()
	// check if the configuration is valid
	if !config.IsValid() {
		clientLoggers.Error.Println("Invalid configuration")
		return
	}
	// connect to the target
	targetConn, err := net.Dial("tcp", config.TargetAddress)
	if err != nil {
		clientLoggers.Error.Println("Error connecting to target:", err)
		return
	}
	// start goroutines to handle data transmission in both directions
	closeChannel := make(chan bool)
	go proxyTargetToClient(closeChannel, conn, targetConn)
	go proxyClientToTarget(closeChannel, conn, targetConn, config)
	// wait for the goroutines to finish
	<-closeChannel
	clientLoggers.Info.Println("Connection of client", conn.RemoteAddr(), "closed")
}
