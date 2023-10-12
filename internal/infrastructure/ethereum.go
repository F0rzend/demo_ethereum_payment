package infrastructure

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient/gethclient"
	"github.com/ethereum/go-ethereum/rpc"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
	"log"
	"math/big"
)

type Ethereum struct {
	client *gethclient.Client
	wallet accounts.Wallet
}

func NewEthereum(ctx context.Context, rpcURL string, walletMnemonic string) (*Ethereum, error) {
	rpcClient, err := rpc.DialContext(ctx, rpcURL)
	if err != nil {
		return nil, fmt.Errorf("cannot dial with ethereum rpc %q: %w", rpcURL, err)
	}

	client := gethclient.New(rpcClient)

	wallet, err := hdwallet.NewFromMnemonic(walletMnemonic)
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet: %w", err)
	}

	return &Ethereum{
		client: client,
		wallet: wallet,
	}, nil
}

func (e *Ethereum) GetInvoiceAccount(id *big.Int) (*common.Address, error) {
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

func (e *Ethereum) generateDerivativePath(id *big.Int) accounts.DerivationPath {
	// TODO: not all id's can be converted to uint32, need to generate subpaths for big amounts of invoices
	path := append(accounts.DefaultBaseDerivationPath, uint32(id.Uint64()))

	return path
}

type TransactionHandler func(*types.Transaction) error

func (e *Ethereum) ListenIncomeTransactions(handler TransactionHandler) {
	transactions := make(chan *types.Transaction)

	// TODO: Using ctx here returns error "context canceled" after first transaction
	sub, err := e.client.SubscribeFullPendingTransactions(context.TODO(), transactions)
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
			go func(handler TransactionHandler, tx *types.Transaction) {
				if err := handler(tx); err != nil {
					log.Printf("failed to handle transaction: %s\n", err)
					return
				}
			}(handler, tx)
		}
	}
}
