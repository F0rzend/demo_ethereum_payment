package transport

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"

	"github.com/F0rzend/demo_ethereum_payment/internal/application"
)

type HTTPHandlers struct {
	application *application.Application
}

func NewHTTPHandlers(
	application *application.Application,
) *HTTPHandlers {
	return &HTTPHandlers{
		application: application,
	}
}

func (s *HTTPHandlers) GetRouter() http.Handler {
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
