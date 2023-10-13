package transport

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"

	"github.com/F0rzend/demo_ethereum_payment/internal/application"
)

const ReadHeaderTimeout = 5 * time.Second

type HTTPServer struct {
	address     string
	application *application.Application
}

func NewHTTPServer(
	address string,
	application *application.Application,
) *HTTPServer {
	return &HTTPServer{
		address:     address,
		application: application,
	}
}

func (s *HTTPServer) ListenAndServe() error {
	server := &http.Server{
		Addr:              s.address,
		ReadHeaderTimeout: ReadHeaderTimeout,
		Handler:           s.getHandler(),
	}

	log.Println("run app")
	if err := server.ListenAndServe(); err != nil {
		return fmt.Errorf("failed to listen and serve: %w", err)
	}

	return nil
}

func (s *HTTPServer) getHandler() http.Handler {
	r := chi.NewRouter()

	r.Use(
		middleware.Recoverer,
		middleware.AllowContentType("application/json"),
		render.SetContentType(render.ContentTypeJSON),
	)

	r.Post("/invoices", ErrorHandler(s.createInvoice))
	r.Get("/invoices/{id}", ErrorHandler(s.getInvoice))

	return r
}
