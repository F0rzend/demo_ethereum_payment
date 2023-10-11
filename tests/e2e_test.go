package tests

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"math/big"
	"net/http"
)

type JSON = map[string]any

func (s *TestSuite) TestApplication() {
	const price = 1

	invoiceID := s.e().
		POST("/invoices").
		WithJSON(JSON{
			"price": price,
		}).
		Expect().
		Status(http.StatusOK).
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

	tx, err := s.testAccount.Transfer(&invoiceAddress, big.NewInt(price))
	require.NoError(s.T(), err)

	s.T().Log(tx.Hash())

	_, err = s.testAccount.WaitForReceipt(tx)
	require.NoError(s.T(), err)

	invoice = s.e().
		GET(invoicePath).
		Expect().
		Status(http.StatusOK).
		JSON().Object()

	invoice.Value("status").String().Equal("paid")
}
