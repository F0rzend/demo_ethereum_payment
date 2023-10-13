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

	gasPrice, err := getOptimalGasPrice(ctx, a.client)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate gas price: %w", err)
	}

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

const priorityCoefficient = 1.25

func getOptimalGasPrice(ctx context.Context, client *ethclient.Client) (*big.Int, error) {
	lastBlockHeader, err := client.HeaderByNumber(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest block header: %w", err)
	}

	baseFee := lastBlockHeader.BaseFee

	maxPriorityFeePerGas, err := client.SuggestGasTipCap(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get priotiry fee: %w", err)
	}

	//maxFeePerGas := big.NewInt(0).Add(
	//	maxPriorityFeePerGas,
	//	baseFee.Mul(baseFee, big.NewInt(priorityCoefficient)),
	//)

	maxFeePerGas, _ := big.NewFloat(0).Add(
		big.NewFloat(0).SetInt(maxPriorityFeePerGas),
		big.NewFloat(0).Mul(big.NewFloat(0).SetInt(baseFee), big.NewFloat(priorityCoefficient)),
	).Int(nil)

	return maxFeePerGas, nil
}

const (
	blockDelay  = 12 * time.Second
	waitTimeout = blockDelay * 10
	waitDelay   = 500 * time.Millisecond
)

func (a *Account) WaitForReceipt(ctx context.Context, tx *types.Transaction) (*types.Receipt, error) {
	ctx, cancel := context.WithTimeout(ctx, waitTimeout)
	defer cancel()

	ticker := time.NewTicker(waitDelay)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf(
				"transaction №%d %s stuck",
				tx.Nonce(),
				tx.Hash().Hex(),
			)
		case <-ticker.C:
			receipt, err := a.client.TransactionReceipt(ctx, tx.Hash())
			if err != nil && !errors.Is(err, ethereum.NotFound) {
				return nil, fmt.Errorf("failed to get receipt: %w", err)
			}

			if receipt == nil {
				continue
			}

			return receipt, nil
		}
	}
}
