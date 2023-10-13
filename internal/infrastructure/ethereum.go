package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts"
	geth "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/ethclient/gethclient"
	"github.com/ethereum/go-ethereum/rpc"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"

	"github.com/F0rzend/demo_ethereum_payment/internal/common"
	"github.com/F0rzend/demo_ethereum_payment/internal/domain"
)

type ethereumClient struct {
	geth *gethclient.Client
	eth  *ethclient.Client
}

type Ethereum struct {
	client *ethereumClient
	wallet accounts.Wallet
}

func NewEthereum(ctx context.Context, rpcURL string, walletMnemonic string) (*Ethereum, error) {
	rpcClient, err := rpc.DialContext(ctx, rpcURL)
	if err != nil {
		return nil, fmt.Errorf("cannot dial with ethereum rpc %q: %w", rpcURL, err)
	}

	client := &ethereumClient{
		geth: gethclient.New(rpcClient),
		eth:  ethclient.NewClient(rpcClient),
	}

	wallet, err := hdwallet.NewFromMnemonic(walletMnemonic)
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet: %w", err)
	}

	return &Ethereum{
		client: client,
		wallet: wallet,
	}, nil
}

func (e *Ethereum) GetInvoiceAccount(id domain.ID) (*geth.Address, error) {
	path := e.generateDerivativePath(id)
	if path.String() == accounts.DefaultRootDerivationPath.String() {
		return nil, fmt.Errorf("you can't use default root derivation path")
	}

	account, err := e.wallet.Derive(path, false)
	if err != nil {
		return nil, fmt.Errorf("failed to derive account with path %s: %w", path, err)
	}

	return &account.Address, nil
}

func (e *Ethereum) generateDerivativePath(id domain.ID) accounts.DerivationPath {
	path := accounts.DefaultRootDerivationPath

	path = append(path, id)

	return path
}

type TransactionHandler func(context.Context, *types.Transaction) error

func (e *Ethereum) ListenIncomeTransactions(ctx context.Context, handler TransactionHandler) {
	transactions := make(chan *types.Transaction)

	sub, err := e.client.geth.SubscribeFullPendingTransactions(ctx, transactions)
	if err != nil {
		log.Printf("failed to subscribe to logs: %s\n", err)

		return
	}

	for {
		select {
		case err := <-sub.Err():
			log.Printf("subscription error: %s\n", err)

			return
		case tx := <-transactions:
			go func(ctx context.Context, handler TransactionHandler, tx *types.Transaction) {
				if err := handler(ctx, tx); err != nil {
					log.Printf("failed to handle transaction: %s\n", err)

					return
				}
			}(ctx, handler, tx)
		}
	}
}

const (
	waitTimeout = 5 * time.Minute
	waitDelay   = 500 * time.Millisecond
)

func (e *Ethereum) WaitForReceipt(ctx context.Context, tx *types.Transaction) (*types.Receipt, error) {
	ctx, cancel := context.WithTimeout(ctx, waitTimeout)
	defer cancel()

	ticker := time.NewTicker(waitDelay)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, common.FlagError(fmt.Errorf("timeout exceeded"), common.FlagTimeout)
		case <-ticker.C:
			receipt, err := e.client.eth.TransactionReceipt(ctx, tx.Hash())
			if err != nil && !errors.Is(err, ethereum.NotFound) {
				return nil, fmt.Errorf("failed to get receipt: %w", err)
			}

			if receipt != nil {
				return receipt, nil
			}
		}
	}
}
