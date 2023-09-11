package configuration

/*
Author: Bastien Faivre
Project: EPFL, DCL, Performance and Security Evaluation of Layer 2 Blockchain Systems
Date: September 2023
*/

import (
	"encoding/json"
	"errors"
	"sync"
)

//------------------------------------------------------------------------------
// Types
//------------------------------------------------------------------------------

type Config struct {
	TargetAddress     string   `json:"targetAddress"`
	CensoredAddresses []string `json:"censoredAddresses"`
}

type ConfigManager struct {
	ConfigLock sync.Mutex
	Config     Config
}

//------------------------------------------------------------------------------
// Errors
//------------------------------------------------------------------------------

// ErrInvalidConfig is returned when the config is invalid
var ErrInvalidConfig = errors.New("invalid config")

//------------------------------------------------------------------------------
// Public methods
//------------------------------------------------------------------------------

// NewConfigManager creates and returns a new ConfigManager
func NewConfigManager(targetAddress string) *ConfigManager {
	return &ConfigManager{
		Config: Config{
			TargetAddress:     targetAddress,
			CensoredAddresses: []string{},
		},
	}
}

// IsValid checks if the config is valid
func (c *Config) IsValid() bool {
	// check that the target address is set
	if c.TargetAddress == "" {
		return false
	}
	// check that the censored addresses array is set
	return c.CensoredAddresses != nil
}

// ParseConfig parses the config from a string
func (cm *ConfigManager) ParseConfig(configStr string) (Config, error) {
	var config Config
	err := json.Unmarshal([]byte(configStr), &config)
	if err != nil {
		return Config{}, err
	}
	if config.TargetAddress == "" {
		config.TargetAddress = cm.Config.TargetAddress
	}
	if !config.IsValid() {
		return Config{}, ErrInvalidConfig
	}
	return config, nil
}

// GetConfig updates the config if it is valid
func (cm *ConfigManager) SetConfig(config Config) error {
	if !config.IsValid() {
		return ErrInvalidConfig
	}
	cm.ConfigLock.Lock()
	defer cm.ConfigLock.Unlock()
	cm.Config = config
	return nil
}

// GetConfig returns the config
func (cm *ConfigManager) GetConfig() Config {
	cm.ConfigLock.Lock()
	defer cm.ConfigLock.Unlock()
	return cm.Config
}

// String returns the string representation of the config
func (c *Config) String() string {
	str := "Config:\n"
	str += "\tTarget Address: " + c.TargetAddress + "\n"
	str += "\tCensored Addresses:\n"
	for _, address := range c.CensoredAddresses {
		str += "\t\t" + address + "\n"
	}
	return str
}
