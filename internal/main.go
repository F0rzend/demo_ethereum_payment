package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/F0rzend/demo_ethereum_payment/internal/application"
	"github.com/F0rzend/demo_ethereum_payment/internal/infrastructure"
	"github.com/F0rzend/demo_ethereum_payment/internal/transport"
)

const (
	MnemonicKey    = "MNEMONIC"
	EthereumRPCKey = "ETHEREUM_RPC"

	ReadHeaderTimeout     = 5 * time.Second
	ServerShutdownTimeout = 60 * time.Second
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

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

	handlers := transport.NewHTTPHandlers(app)

	server := &http.Server{
		Addr:              ":8080",
		ReadHeaderTimeout: ReadHeaderTimeout,
		Handler:           handlers.GetRouter(),
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		err := app.StartListeningTransactions(ctx)
		if err != nil {
			return fmt.Errorf("failed to start listening transactions: %w", err)
		}

		return nil
	})
	g.Go(func() error {
		log.Println("server started")
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("failed to start listening and serving: %w", err)
		}

		return nil
	})
	g.Go(func() error {
		<-ctx.Done()

		log.Println("server stopping")
		ctx, cancel := context.WithTimeout(context.Background(), ServerShutdownTimeout)
		defer cancel()

		//nolint:contextcheck
		if err := server.Shutdown(ctx); err != nil {
			return fmt.Errorf("failed to shutdown server: %w", err)
		}

		log.Println("server stopped")

		return nil
	})

	if err := g.Wait(); err != nil {
		return fmt.Errorf("failed to wait group: %w", err)
	}

	return nil
}
