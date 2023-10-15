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

// ShutdownOnContextDone returns a child context and a function that will shut the server down
// when the signalContext is done.
// After the server is shutdown, the child context will be canceled.
// Use child context to run goroutines that should be stopped when the server is shutdown.
func (s *HTTPServer) ShutdownOnContextDone(
	signalContext context.Context,
) (context.Context, common.ErrorGroupGoroutine) {
	childContext, cancelChildContexts := context.WithCancel(context.Background())

	return childContext, func() error {
		defer cancelChildContexts()

		<-signalContext.Done()

		ctx, cancel := context.WithTimeout(childContext, ServerShutdownTimeout)
		defer cancel()

		if err := s.server.Shutdown(ctx); err != nil {
			return fmt.Errorf("failed to shutdown server: %w", err)
		}

		log.Println("server stopped")

		return nil
	}
}
