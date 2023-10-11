package tests

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
	"math/big"
	"time"
)

type Account struct {
	address common.Address
	private *ecdsa.PrivateKey
	public  *ecdsa.PublicKey

	client  *ethclient.Client
	chainID *big.Int
}

func NewAccount(rpcURL string, rawPrivate string) (*Account, error) {
	ctx := context.Background()

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

	log.Println("suggested gas price:", suggestedGasPrice)

	gasPrice := getOptimalGasPrice(suggestedGasPrice)

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

// getOptimalGasPrice returns gas price that is 25% higher than suggested.
func getOptimalGasPrice(suggested *big.Int) *big.Int {
	return big.NewInt(0).Add(
		suggested,
		big.NewInt(0).Div(suggested, big.NewInt(4)),
	)
}

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

		time.Sleep(500 * time.Millisecond)
	}
}
