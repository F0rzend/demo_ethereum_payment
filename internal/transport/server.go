package transport

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/F0rzend/demo_ethereum_payment/internal/application"
	"github.com/F0rzend/demo_ethereum_payment/internal/common"
)

const (
	ReadHeaderTimeout     = 5 * time.Second
	ServerShutdownTimeout = 60 * time.Second
)

type HTTPServer struct {
	server *http.Server
}

func NewHTTPServer(ctx context.Context, address string, app *application.Application) *HTTPServer {
	handlers := NewHTTPHandlers(app)

	server := &http.Server{
		Addr:              address,
		ReadHeaderTimeout: ReadHeaderTimeout,
		Handler:           handlers.GetRouter(),
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}

	return &HTTPServer{
		server: server,
	}
}

func (s *HTTPServer) Run() error {
	log.Println("server started")

	err := s.server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("failed to start listening and serving: %w", err)
	}

	return nil
}

func (s *HTTPServer) ShutdownOnContextDone(ctx context.Context) common.ErrorGroupGoroutine {
	return func() error {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), ServerShutdownTimeout)
		defer cancel()

		//nolint:contextcheck
		if err := s.server.Shutdown(ctx); err != nil {
			return fmt.Errorf("failed to shutdown server: %w", err)
		}

		log.Println("server stopped")

		return nil
	}
}
