package application

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"sync"

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

func (a *Application) RunTransactionHandler(ctx context.Context) common.ErrorGroupGoroutine {
	return func() error {
		transactions, err := a.ethereum.SubscribeConfirmedTransactions(ctx)
		if err != nil {
			return fmt.Errorf("failed to subscribe to confirmed transactions: %w", err)
		}

		log.Println("start handling new confirmed transactions")
		wg := new(sync.WaitGroup)
		for tx := range transactions {
			wg.Add(1)
			go func(tx *types.Transaction) {
				defer wg.Done()

				if err := a.handleTransaction(tx); err != nil {
					log.Printf("failed to handle transaction: %s\n", err)
				}
			}(tx)
		}

		wg.Wait()

		log.Println("handling new confirmed transactions stopped")

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
