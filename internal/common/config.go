package common

import (
	"fmt"
	"os"
)

type Config struct {
	Mnemonic      string
	EthereumRPC   string
	ServerAddress string
}

const (
	MnemonicKey      = "MNEMONIC"
	EthereumRPCKey   = "ETHEREUM_RPC"
	ServerAddressKey = "SERVER_ADDRESS"
)

func ConfigFromEnv() (*Config, error) {
	mnemonic, ok := os.LookupEnv(MnemonicKey)
	if !ok {
		return nil, fmt.Errorf("environment variable %s not set", MnemonicKey)
	}

	ethereumRPC, ok := os.LookupEnv(EthereumRPCKey)
	if !ok {
		return nil, fmt.Errorf("environment variable %s not set", EthereumRPCKey)
	}

	serverAddress, ok := os.LookupEnv(ServerAddressKey)
	if !ok {
		return nil, fmt.Errorf("environment variable %s not set", ServerAddressKey)
	}

	return &Config{
		Mnemonic:      mnemonic,
		EthereumRPC:   ethereumRPC,
		ServerAddress: serverAddress,
	}, nil
}
