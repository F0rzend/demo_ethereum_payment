package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"

	"github.com/F0rzend/demo_ethereum_payment/internal/application"
	"github.com/F0rzend/demo_ethereum_payment/internal/common"
	"github.com/F0rzend/demo_ethereum_payment/internal/infrastructure"
	"github.com/F0rzend/demo_ethereum_payment/internal/transport"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	config, err := common.ConfigFromEnv()
	if err != nil {
		return fmt.Errorf("cannot get config from env: %w", err)
	}

	ethereum, err := infrastructure.NewEthereum(ctx, config.EthereumRPC, config.Mnemonic)
	if err != nil {
		return fmt.Errorf("cannot create ethereum gataway: %w", err)
	}

	repository := infrastructure.NewRepository()
	app := application.NewApplication(ethereum, repository)

	server := transport.NewHTTPServer(ctx, config.ServerAddress, app)

	g, ctx := errgroup.WithContext(ctx)

	ctx, shutdownFn := server.ShutdownOnContextDone(ctx)

	g.Go(shutdownFn)
	g.Go(app.RunTransactionHandler(ctx))
	g.Go(server.Run)

	if err := g.Wait(); err != nil {
		return fmt.Errorf("an unexpected error occurred while the application was running: %w", err)
	}

	log.Println("server stopped gracefully")

	return nil
}
