package infrastructure

import (
	"fmt"
	"sync"

	geth "github.com/ethereum/go-ethereum/common"

	"github.com/F0rzend/demo_ethereum_payment/internal/common"
	"github.com/F0rzend/demo_ethereum_payment/internal/domain"
)

type Repository struct {
	invoices       *sync.Map
	addressesIndex *sync.Map

	lastID domain.ID

	mu *sync.Mutex
}

func NewRepository() *Repository {
	return &Repository{
		invoices:       new(sync.Map),
		addressesIndex: new(sync.Map),
		lastID:         0,
		mu:             &sync.Mutex{},
	}
}

func (r *Repository) GetID() domain.ID {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.lastID++

	return r.lastID
}

func (r *Repository) Save(invoice *domain.Invoice) {
	r.invoices.Store(invoice.ID(), invoice)
	r.addressesIndex.Store(invoice.Address().Hex(), invoice)
}

func (r *Repository) GetByID(id domain.ID) (*domain.Invoice, error) {
	invoice, ok := r.invoices.Load(id)
	if !ok {
		return nil, common.FlagError(fmt.Errorf("invoice with id %d not found", id), common.FlagNotFound)
	}

	typedInvoice, ok := invoice.(*domain.Invoice)
	if !ok {
		return nil, fmt.Errorf("invoice with id %d has invalid type", id)
	}

	return typedInvoice, nil
}

func (r *Repository) GetByAddress(address *geth.Address) (*domain.Invoice, error) {
	value, ok := r.addressesIndex.Load(address.Hex())
	if !ok {
		return nil, common.FlagError(
			fmt.Errorf("invoice with address %q not found", address.Hex()),
			common.FlagNotFound,
		)
	}

	invoice, ok := value.(*domain.Invoice)
	if !ok {
		return nil, fmt.Errorf("invoice with address %q has invalid type", address.Hex())
	}

	return invoice, nil
}
