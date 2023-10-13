package tests

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	LowGasPriceCoefficient      = 1
	MarketGasFeeCoefficient     = 1.25
	AggressiveGasFeeCoefficient = 1.5
)

type Account struct {
	address common.Address
	private *ecdsa.PrivateKey
	public  *ecdsa.PublicKey

	client  *ethclient.Client
	chainID *big.Int
}

func NewAccount(ctx context.Context, rpcURL string, rawPrivate string) (*Account, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ethereum node: %w", err)
	}

	chainID, err := client.NetworkID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get network id: %w", err)
	}

	privateKey, err := crypto.HexToECDSA(rawPrivate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("failed to cast public key to ECDSA")
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	return &Account{
		address: address,
		private: privateKey,
		public:  publicKeyECDSA,
		client:  client,
		chainID: chainID,
	}, nil
}

func (a *Account) Transfer(to *common.Address, value *big.Int) (*types.Transaction, error) {
	ctx := context.Background()

	nonce, err := a.client.PendingNonceAt(ctx, a.address)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %w", err)
	}

	suggestedGasPrice, err := a.client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get suggested gas price: %w", err)
	}

	gasPrice := getOptimalGasPrice(suggestedGasPrice, MarketGasFeeCoefficient)

	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		GasPrice: gasPrice,
		Gas:      TransferGasLimit,
		To:       to,
		Value:    value,
		Data:     nil,
	})

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(a.chainID), a.private)
	if err != nil {
		return nil, fmt.Errorf("failed to sign tx: %w", err)
	}

	err = a.client.SendTransaction(ctx, signedTx)
	if err != nil {
		return nil, fmt.Errorf("failed to send tx: %w", err)
	}

	return signedTx, nil
}

func getOptimalGasPrice(suggested *big.Int, coefficient float64) *big.Int {
	value := big.NewFloat(0).SetInt(suggested)

	result, _ := value.
		Mul(value, big.NewFloat(coefficient)).
		Int(nil)

	return result
}

const waitDelay = 500 * time.Millisecond

func (a *Account) WaitForReceipt(tx *types.Transaction) (*types.Receipt, error) {
	ctx := context.Background()

	for {
		receipt, err := a.client.TransactionReceipt(ctx, tx.Hash())
		if err != nil && !errors.Is(err, ethereum.NotFound) {
			return nil, fmt.Errorf("failed to get receipt: %w", err)
		}

		if receipt != nil {
			return receipt, nil
		}

		time.Sleep(waitDelay)
	}
}
