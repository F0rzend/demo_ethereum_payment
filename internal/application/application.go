package application

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"

	"github.com/F0rzend/demo_ethereum_payment/internal/common"
	"github.com/F0rzend/demo_ethereum_payment/internal/domain"
	"github.com/F0rzend/demo_ethereum_payment/internal/infrastructure"
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

func (a *Application) RunTransactionListener(ctx context.Context) common.ErrorGroupGoroutine {
	return func() error {
		if err := a.ethereum.ListenConfirmedTransactions(ctx, a.handleTransaction); err != nil {
			return fmt.Errorf("failed to start listening transactions: %w", err)
		}

		return nil
	}
}

func (a *Application) CreateInvoice(price domain.WEI) (domain.ID, error) {
	id := a.repository.GetID()

	invoiceAddress, err := a.ethereum.GetInvoiceAccount(id)
	if err != nil {
		return 0, fmt.Errorf("failed to get invoice account: %w", err)
	}

	invoice := domain.NewInvoice(
		id,
		price,
		big.NewInt(0),
		invoiceAddress,
		domain.InvoiceStatusPending,
	)

	a.repository.Save(invoice)

	return invoice.ID(), nil
}

func (a *Application) handleTransaction(tx *types.Transaction) error {
	if tx.To() == nil {
		return nil
	}

	value := tx.Value()

	invoice, err := a.repository.GetByAddress(tx.To())
	if common.IsFlaggedError(err, common.FlagNotFound) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("cannot get invoice by address: %w", err)
	}

	invoice.Deposit(value)

	a.repository.Save(invoice)

	return nil
}

func (a *Application) GetInvoice(id domain.ID) (*domain.Invoice, error) {
	invoice, err := a.repository.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}

	return invoice, nil
}
