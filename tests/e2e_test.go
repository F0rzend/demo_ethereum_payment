package tests

import (
	"context"
	"fmt"
	"math/big"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

type JSON = map[string]any

func (s *TestSuite) TestApplication() {
	ctx := context.Background()

	const transactionValue = 1

	invoiceID := s.e().
		POST("/invoices").
		WithJSON(JSON{
			"price": transactionValue * 2,
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object().
		Value("id").Number().Raw()

	invoicePath := fmt.Sprintf("/invoices/%d", int(invoiceID))
	invoice := s.e().
		GET(invoicePath).
		Expect().
		Status(http.StatusOK).
		JSON().Object()

	invoice.Value("status").String().Equal("pending")

	invoiceAddressHex := invoice.Value("address").String().Raw()
	invoiceAddress := common.HexToAddress(invoiceAddressHex)

	tx, err := s.testAccount.Transfer(&invoiceAddress, big.NewInt(transactionValue))
	require.NoError(s.T(), err)

	s.T().Logf("waiting for transaction %s", tx.Hash().Hex())

	receipt, err := s.testAccount.WaitForReceipt(ctx, tx)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), receipt)

	invoice = s.e().
		GET(invoicePath).
		Expect().
		Status(http.StatusOK).
		JSON().Object()

	invoice.Value("status").String().Equal("pending")

	tx, err = s.testAccount.Transfer(&invoiceAddress, big.NewInt(transactionValue))
	require.NoError(s.T(), err)

	s.T().Logf("waiting for transaction %s", tx.Hash().Hex())

	receipt, err = s.testAccount.WaitForReceipt(ctx, tx)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), receipt)

	invoice = s.e().
		GET(invoicePath).
		Expect().
		Status(http.StatusOK).
		JSON().Object()

	invoice.Value("status").String().Equal("paid")
}
