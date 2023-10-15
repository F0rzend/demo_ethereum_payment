package infrastructure

import (
	"context"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts"
	geth "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"

	"github.com/F0rzend/demo_ethereum_payment/internal/domain"
)

type Ethereum struct {
	client *ethclient.Client
	wallet accounts.Wallet
}

func NewEthereum(ctx context.Context, rpcURL string, walletMnemonic string) (*Ethereum, error) {
	client, err := ethclient.DialContext(ctx, rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to dial to %s: %w", rpcURL, err)
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

type TransactionHandler func(*types.Transaction) error

func (e *Ethereum) SubscribeConfirmedTransactions(ctx context.Context) (<-chan *types.Transaction, error) {
	headers := make(chan *types.Header)

	sub, err := e.client.SubscribeNewHead(ctx, headers)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to headers: %w", err)
	}

	transactions := make(chan *types.Transaction)

	go e.listenHeaders(ctx, headers, sub, transactions)

	return transactions, nil
}

func (e *Ethereum) listenHeaders(
	ctx context.Context,
	headers <-chan *types.Header,
	headersSubscription ethereum.Subscription,
	transactions chan<- *types.Transaction,
) {
	defer close(transactions)

	for {
		select {
		case <-ctx.Done():
			headersSubscription.Unsubscribe()
			log.Println("unsubscribed from headers")

			return
		case err := <-headersSubscription.Err():
			log.Printf("subscription error: %s\n", err)

			return
		case header := <-headers:
			block, err := e.client.BlockByHash(ctx, header.Hash())
			if err != nil {
				log.Printf("failed to get block by hash %s: %s\n", header.Hash(), err)

				continue
			}

			blockTransactions := block.Transactions()

			log.Printf("new block received with %d transactions\n", len(blockTransactions))
			for _, tx := range blockTransactions {
				transactions <- tx
			}
		}
	}
}
