package main

import (
	"context"
	"ethereum_payment_demo/internal/application"
	"ethereum_payment_demo/internal/delivery"
	"ethereum_payment_demo/internal/infrastructure"
	"fmt"
	"log"
	"os"
)

const (
	MnemonicKey    = "MNEMONIC"
	EthereumRPCKey = "ETHEREUM_RPC"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	mnemonic, ok := os.LookupEnv(MnemonicKey)
	if !ok {
		return fmt.Errorf("environment variable %s not set", MnemonicKey)
	}

	ethereumRPC, ok := os.LookupEnv(EthereumRPCKey)
	if !ok {
		return fmt.Errorf("environment variable %s not set", EthereumRPCKey)
	}

	repository := infrastructure.NewRepository()
	ethereum, err := infrastructure.NewEthereum(context.Background(), ethereumRPC, mnemonic)
	if err != nil {
		return fmt.Errorf("cannot create ethereum gataway: %w", err)
	}

	app := application.NewApplication(ethereum, repository)
	server := delivery.NewHTTPServer(":8080", app)

	log.Println("run app")
	return server.ListenAndServe()
}
