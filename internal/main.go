package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/F0rzend/demo_ethereum_payment/internal/application"
	"github.com/F0rzend/demo_ethereum_payment/internal/infrastructure"
	"github.com/F0rzend/demo_ethereum_payment/internal/transport"
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
	ctx := context.Background()

	mnemonic, ok := os.LookupEnv(MnemonicKey)
	if !ok {
		return fmt.Errorf("environment variable %s not set", MnemonicKey)
	}

	ethereumRPC, ok := os.LookupEnv(EthereumRPCKey)
	if !ok {
		return fmt.Errorf("environment variable %s not set", EthereumRPCKey)
	}

	repository := infrastructure.NewRepository()

	ethereum, err := infrastructure.NewEthereum(ctx, ethereumRPC, mnemonic)
	if err != nil {
		return fmt.Errorf("cannot create ethereum gataway: %w", err)
	}

	app := application.NewApplication(ethereum, repository)
	go app.StartListeningTransactions(ctx)

	server := transport.NewHTTPServer(":8080", app)

	if err := server.ListenAndServe(); err != nil {
		return fmt.Errorf("failed to listen and serve: %w", err)
	}

	return nil
}
