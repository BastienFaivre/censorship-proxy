package main

/*
Author: Bastien Faivre
Project: EPFL, DCL, Performance and Security Evaluation of Layer 2 Blockchain
		 Systems
Date: September 2023
*/

import (
	"fmt"
	"io"
	"net"
	"os"
)

func handleConnection(clientConn net.Conn, targetAddress string) {
	defer clientConn.Close()
	// connect to target
	targetConn, err := net.Dial("tcp", targetAddress)
	if err != nil {
		fmt.Println("Error connecting to target: " + err.Error())
		return
	}
	defer targetConn.Close()
	fmt.Println("Connected to target " + targetConn.RemoteAddr().String())
	// forward data
	closeChannel := make(chan bool)
	go func() {
		io.Copy(targetConn, clientConn)
		closeChannel <- true
	}()
	go func() {
		io.Copy(clientConn, targetConn)
		closeChannel <- true
	}()
	<-closeChannel
	fmt.Println("Connection closed")
}

func listen(listeningAddress string, targetAddress string) {
	listener, err := net.Listen("tcp", listeningAddress)
	if err != nil {
		panic("Error listening: " + err.Error())
	}
	defer listener.Close()
	// listen for incoming connections
	for {
		clientConn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting new connection: " + err.Error())
			continue
		}
		fmt.Println("New connection from " + clientConn.RemoteAddr().String())
		// handle connections in a new goroutine
		go handleConnection(clientConn, targetAddress)
	}
}

func main() {
	// read arguments
	if len(os.Args) != 3 {
		fmt.Println("Usage: proxy <listening address> <target address>")
		os.Exit(1)
	}
	listeningAddress := os.Args[1]
	targetAddress := os.Args[2]
	// listen
	listen(listeningAddress, targetAddress)
}
