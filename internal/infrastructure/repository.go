package infrastructure

import (
	"ethereum_payment_demo/internal/domain"
	"fmt"
	"math/big"
	"sync"
)

type Repository struct {
	invoices *sync.Map
	lastID   *big.Int

	mu *sync.Mutex
}

func NewRepository() *Repository {
	return &Repository{
		invoices: &sync.Map{},
		lastID:   big.NewInt(0),
		mu:       &sync.Mutex{},
	}
}

func (r *Repository) GetID() *big.Int {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.lastID.Add(r.lastID, big.NewInt(1))

	return r.lastID
}

func (r *Repository) Save(invoice *domain.Invoice) {
	r.invoices.Store(invoice.ID().String(), invoice)
}

func (r *Repository) Get(id *big.Int) (*domain.Invoice, error) {
	invoice, ok := r.invoices.Load(id.String())
	if !ok {
		return nil, fmt.Errorf("invoice with id %s not found", id)
	}

	typedInvoice, ok := invoice.(*domain.Invoice)
	if !ok {
		return nil, fmt.Errorf("invoice with id %s has invalid type", id)
	}

	return typedInvoice, nil
}
