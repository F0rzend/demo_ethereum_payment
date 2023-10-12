package infrastructure

import (
	"ethereum_payment_demo/internal/common"
	"ethereum_payment_demo/internal/domain"
	"fmt"
	eth "github.com/ethereum/go-ethereum/common"
	"math/big"
	"sync"
)

type Repository struct {
	invoices       *sync.Map
	addressesIndex *sync.Map

	lastID *big.Int

	mu *sync.Mutex
}

func NewRepository() *Repository {
	return &Repository{
		invoices:       new(sync.Map),
		addressesIndex: new(sync.Map),
		lastID:         big.NewInt(0),
		mu:             &sync.Mutex{},
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
	r.addressesIndex.Store(invoice.Address().Hex(), invoice)
}

func (r *Repository) GetByID(id domain.ID) (*domain.Invoice, error) {
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

func (r *Repository) GetByAddress(address *eth.Address) (*domain.Invoice, error) {
	value, ok := r.addressesIndex.Load(address.Hex())
	if !ok {
		return nil, common.FlagError(
			fmt.Errorf("invoice with address %q not found", address.Hex()),
			common.NotExistsFlag,
		)
	}

	invoice, ok := value.(*domain.Invoice)
	if !ok {
		return nil, fmt.Errorf("invoice with address %q has invalid type", address.Hex())
	}

	return invoice, nil
}
