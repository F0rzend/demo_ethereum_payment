package infrastructure

import (
	"context"
	"fmt"
	"log"
	"sync"

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

func (e *Ethereum) ListenConfirmedTransactions(ctx context.Context, handler TransactionHandler) error {
	headers := make(chan *types.Header)

	sub, err := e.client.SubscribeNewHead(ctx, headers)
	if err != nil {
		return fmt.Errorf("failed to subscribe to headers: %w", err)
	}

	wg := new(sync.WaitGroup)
	defer wg.Wait()

	for {
		select {
		case <-ctx.Done():
			sub.Unsubscribe()
			log.Println("unsubscribed from headers")

			return nil
		case err := <-sub.Err():
			return fmt.Errorf("subscription error: %w", err)
		case header := <-headers:
			if err := e.processHeader(ctx, wg, header, handler); err != nil {
				return fmt.Errorf("failed to process header: %w", err)
			}
		}
	}
}

func (e *Ethereum) processHeader(
	ctx context.Context,
	handlersWaitGroup *sync.WaitGroup,
	header *types.Header,
	handler TransactionHandler,
) error {
	block, err := e.client.BlockByHash(ctx, header.Hash())
	if err != nil {
		return fmt.Errorf("failed to get block by hash %s: %w", header.Hash(), err)
	}

	handlersWaitGroup.Add(len(block.Transactions()))
	for _, tx := range block.Transactions() {
		go func(tx *types.Transaction) {
			defer handlersWaitGroup.Done()

			if err := handler(tx); err != nil {
				log.Printf("failed to handle transaction %s: %s\n", tx.Hash().Hex(), err)
			}
		}(tx)
	}

	return nil
}
