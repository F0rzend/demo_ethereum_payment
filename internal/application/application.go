package application

import (
	"ethereum_payment_demo/internal/domain"
	"ethereum_payment_demo/internal/infrastructure"
	"fmt"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"
)

type Application struct {
	ethereum   *infrastructure.Ethereum
	repository *infrastructure.Repository
}

func NewApplication(
	ethereum *infrastructure.Ethereum,
	repository *infrastructure.Repository,
) *Application {
	return &Application{
		ethereum:   ethereum,
		repository: repository,
	}
}

func (a *Application) CreateInvoice(price domain.WEI) (domain.ID, error) {
	id := a.repository.GetID()

	invoiceAddress, err := a.ethereum.GetInvoiceAccount(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice account: %w", err)
	}

	invoice := domain.NewInvoice(
		id,
		price,
		big.NewInt(0),
		invoiceAddress,
		domain.InvoiceStatusPending,
	)

	a.repository.Save(invoice)
	go a.ethereum.ListenIncomeTransactions(invoice, a.handleTransaction)

	return invoice.ID(), nil
}

func (a *Application) handleTransaction(invoice *domain.Invoice, tx *types.Transaction) error {
	value := tx.Value()

	invoice.Deposit(value)

	a.repository.Save(invoice)

	return nil
}

func (a *Application) GetInvoice(id domain.ID) (*domain.Invoice, error) {
	invoice, err := a.repository.Get(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}

	return invoice, nil
}
