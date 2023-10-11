package delivery

import (
	"encoding/json"
	"ethereum_payment_demo/internal/application"
	"ethereum_payment_demo/internal/domain"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log"
	"math/big"
	"net/http"
)

type HTTPServer struct {
	server      *http.Server
	application *application.Application
}

func NewHTTPServer(
	address string,
	application *application.Application,
) *HTTPServer {
	server := &http.Server{
		Addr: address,
	}

	return &HTTPServer{
		server:      server,
		application: application,
	}
}

func (s *HTTPServer) registerRoutes() {
	r := chi.NewRouter()

	r.Use(
		middleware.Recoverer,
		middleware.AllowContentType("application/json"),
		render.SetContentType(render.ContentTypeJSON),
	)

	r.Post("/invoices", s.createInvoice)
	r.Get("/invoices/{id}", s.getInvoice)

	s.server.Handler = r
}

func (s *HTTPServer) ListenAndServe() error {
	s.registerRoutes()

	err := s.server.ListenAndServe()
	if err != nil {
		return fmt.Errorf("failed to listen and serve: %w", err)
	}

	return nil
}

func (s *HTTPServer) createInvoice(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Price domain.WEI `json:"price"`
	}

	var req request

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("failed to decode request: %s\n", err)
		return
	}

	id, err := s.application.CreateInvoice(req.Price)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("failed to create invoice: %s\n", err)
		return
	}

	type response struct {
		ID domain.ID `json:"id"`
	}

	resp := response{
		ID: id,
	}

	render.Status(r, http.StatusOK)
	render.Respond(w, r, resp)
}

func (s *HTTPServer) getInvoice(w http.ResponseWriter, r *http.Request) {
	rawID := chi.URLParam(r, "id")
	id, ok := big.NewInt(0).SetString(rawID, 10)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("failed to parse id: %s\n", rawID)
		return
	}

	invoice, err := s.application.GetInvoice(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("failed to get invoice: %s\n", err)
		return
	}

	type response struct {
		ID      domain.ID            `json:"id"`
		Price   domain.WEI           `json:"price"`
		Balance domain.WEI           `json:"balance"`
		Address domain.Address       `json:"address"`
		Status  domain.InvoiceStatus `json:"status"`
	}

	resp := response{
		ID:      invoice.ID(),
		Price:   invoice.Price(),
		Balance: invoice.Balance(),
		Address: invoice.Address(),
		Status:  invoice.Status(),
	}

	render.Status(r, http.StatusOK)
	render.Respond(w, r, resp)
}
