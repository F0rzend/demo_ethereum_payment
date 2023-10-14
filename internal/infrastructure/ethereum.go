package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
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

type TransactionHandler func(*types.Transaction) error

func (e *Ethereum) ListenConfirmedTransactions(ctx context.Context, handler TransactionHandler) error {
	headers := make(chan *types.Header)

	sub, err := e.client.eth.SubscribeNewHead(ctx, headers)
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
			block, err := e.client.eth.BlockByHash(ctx, header.Hash())
			if err != nil {
				log.Printf("failed to get block by hash %s: %s\n", header.Hash(), err)

				continue
			}

			wg.Add(len(block.Transactions()))
			for _, tx := range block.Transactions() {
				go func(tx *types.Transaction) {
					if err := handler(tx); err != nil {
						log.Printf("failed to handle transaction %s: %s\n", tx.Hash().Hex(), err)
					}
					wg.Done()
				}(tx)
			}
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
