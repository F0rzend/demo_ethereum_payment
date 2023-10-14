package transport

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	"github.com/F0rzend/demo_ethereum_payment/internal/common"
	"github.com/F0rzend/demo_ethereum_payment/internal/domain"
)

func (s *HTTPHandlers) createInvoice(w http.ResponseWriter, r *http.Request) error {
	type request struct {
		Price domain.WEI `json:"price"`
	}

	var req request

	if err := render.Decode(r, &req); err != nil {
		return NewValidationError("invalid request body")
	}

	id, err := s.application.CreateInvoice(req.Price)
	if err != nil {
		return fmt.Errorf("failed to create invoice: %w", err)
	}

	type response struct {
		ID domain.ID `json:"id"`
	}

	resp := response{
		ID: id,
	}

	render.Status(r, http.StatusCreated)
	render.Respond(w, r, resp)

	return nil
}

const (
	InvoiceIDNumberSystem = 10
	InvoiceIDBitSize      = 32
)

func (s *HTTPHandlers) getInvoice(w http.ResponseWriter, r *http.Request) error {
	rawID := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(rawID, InvoiceIDNumberSystem, InvoiceIDBitSize)
	if err != nil {
		return NewValidationError(
			fmt.Sprintf("failed to parse invoice id %q", rawID),
		)
	}

	invoice, err := s.application.GetInvoice(domain.ID(id))
	if common.IsFlaggedError(err, common.FlagNotFound) {
		return NewNotFoundError(
			fmt.Sprintf("invoice with id %d not found", id),
		)
	}
	if err != nil {
		return fmt.Errorf("failed to get invoice: %w", err)
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

	return nil
}
