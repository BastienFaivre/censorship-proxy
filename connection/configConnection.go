package connection

/*
Author: Bastien Faivre
Project: EPFL, DCL, Performance and Security Evaluation of Layer 2 Blockchain Systems
Date: September 2023
*/

import (
	"censorship-proxy/configuration"
	"censorship-proxy/logs"
	"net"
)

//------------------------------------------------------------------------------
// Constants
//------------------------------------------------------------------------------

// size of the buffer used to read data
const CONFIG_BUFFER_SIZE = 1024

//------------------------------------------------------------------------------
// Private variables
//------------------------------------------------------------------------------

// logger used by the config connection handler
var configLoggers *logs.Loggers

//------------------------------------------------------------------------------
// Public methods
//------------------------------------------------------------------------------

// InitConfigLoggers initializes the loggers for the config connection handler
func InitConfigLoggers(loggers *logs.Loggers) {
	configLoggers = loggers
}

// HandleConfigConnection handles a new configuration connection
func HandleConfigConnection(conn net.Conn, configManager *configuration.ConfigManager) {
	defer conn.Close()
	// read configuration
	data := make([]byte, CONFIG_BUFFER_SIZE)
	n, err := conn.Read(data)
	if err != nil {
		configLoggers.Error.Println("Error reading data:", err)
		return
	}
	// parse configuration
	config, err := configManager.ParseConfig(string(data[:n]))
	if err != nil {
		configLoggers.Error.Println("Error parsing config:", err)
		return
	}
	// set the configuration
	err = configManager.SetConfig(config)
	if err != nil {
		configLoggers.Error.Println("Error setting config:", err)
		return
	}
	configLoggers.Info.Println("Configurations successfully updated")
	configLoggers.Info.Println(config.String())
}
