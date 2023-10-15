package tests

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/gavv/httpexpect"
	"github.com/stretchr/testify/require"
)

type JSON = map[string]any

const (
	InvoiceStatusPending = "pending"
	InvoiceStatusPaid    = "paid"
)

type TestInvoice struct {
	t *testing.T

	ID      *httpexpect.Number
	Price   *httpexpect.Number
	Balance *httpexpect.Number
	Address *httpexpect.String
	Status  *httpexpect.String
}

func (i *TestInvoice) Deposit(from *EthereumGateway, value uint64) *TestTransaction {
	i.t.Helper()

	invoiceAddressHex := i.Address.Raw()
	invoiceAddress := common.HexToAddress(invoiceAddressHex)

	tx, err := from.Transfer(&invoiceAddress, big.NewInt(int64(value)))
	require.NoError(i.t, err)

	i.t.Logf("sending %d wei to invoice#%d with address %s", value, uint64(i.ID.Raw()), invoiceAddressHex)

	return &TestTransaction{
		t:           i.t,
		transaction: tx,
	}
}

type TestTransaction struct {
	t           *testing.T
	transaction *types.Transaction
}

func (t *TestTransaction) WaitConfirmation(ctx context.Context, gateway *EthereumGateway) {
	t.t.Helper()

	receipt, err := gateway.WaitForReceipt(ctx, t.transaction)
	require.NoError(t.t, err)
	require.NotNil(t.t, receipt)
}
