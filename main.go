package main

/*
Author: Bastien Faivre
Project: EPFL, DCL, Performance and Security Evaluation of Layer 2 Blockchain Systems
Date: September 2023
*/

import (
	"censorship-proxy/configuration"
	"censorship-proxy/connection"
	"censorship-proxy/logs"
	"fmt"
	"net"
	"os"
)

// configListener listens for config updates
func configListener(loggers *logs.Loggers, configAddress string, configManager *configuration.ConfigManager) {
	listener, err := net.Listen("tcp", configAddress)
	if err != nil {
		loggers.Error.Println("Error listening for config updates:", err)
		panic(err)
	}
	defer listener.Close()
	loggers.Info.Println("Listening for config updates on", configAddress)
	// listen for connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			loggers.Error.Println("Error accepting new connection:", err)
		}
		loggers.Info.Println("New connection from", conn.RemoteAddr())
		// handle the connection
		go connection.HandleConfigConnection(conn, configManager)
	}
}

// clientListener listens for client connections
func clientListener(loggers *logs.Loggers, clientAddress string, configManager *configuration.ConfigManager) {
	listener, err := net.Listen("tcp", clientAddress)
	if err != nil {
		loggers.Error.Println("Error listening for client connections:", err)
		panic(err)
	}
	defer listener.Close()
	loggers.Info.Println("Listening for client connections on", clientAddress)
	// listen for connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			loggers.Error.Println("Error accepting new connection:", err)
		}
		loggers.Info.Println("New connection from", conn.RemoteAddr())
		// handle the connection
		go connection.HandleClientConnection(conn, configManager)
	}
}

func main() {
	// read arguments
	if len(os.Args) != 4 {
		fmt.Println("Usage: proxy <config address> <client address> <target address>")
		os.Exit(1)
	}
	configAddress := os.Args[1]
	clientAddress := os.Args[2]
	targetAddress := os.Args[3]
	// get loggers
	clientLoggers, configLoggers, err := logs.GetLoggers("")
	if err != nil {
		panic("Error getting loggers: " + err.Error())
	}
	// initialize loggers
	connection.InitClientLoggers(clientLoggers)
	connection.InitConfigLoggers(configLoggers)
	// create a configuration manager
	configManager := configuration.NewConfigManager(targetAddress)
	// listen for config updates
	go configListener(configLoggers, configAddress, configManager)
	// listen for client connections
	clientListener(clientLoggers, clientAddress, configManager)
}
