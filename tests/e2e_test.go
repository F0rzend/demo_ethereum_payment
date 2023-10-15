package tests

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/gavv/httpexpect"
)

func (s *TestSuite) TestApplication() {
	ctx := context.Background()

	const transactionValue = 1

	invoiceID := sendCreateInvoiceRequest(s.T(), s.e(), transactionValue*2)

	invoice := sendGetInvoiceRequest(s.T(), s.e(), invoiceID)
	invoice.Status.Equal(InvoiceStatusPending)

	tx := invoice.Deposit(s.eth, transactionValue)
	tx.WaitConfirmation(ctx, s.eth)
	waitForProcessing(s.T())

	invoice = sendGetInvoiceRequest(s.T(), s.e(), invoiceID)
	invoice.Status.Equal(InvoiceStatusPending)

	tx = invoice.Deposit(s.eth, transactionValue)
	tx.WaitConfirmation(ctx, s.eth)
	waitForProcessing(s.T())

	invoice = sendGetInvoiceRequest(s.T(), s.e(), invoiceID)
	invoice.Status.Equal(InvoiceStatusPaid)
}

func sendCreateInvoiceRequest(t *testing.T, e *httpexpect.Expect, price uint64) uint64 {
	t.Helper()

	invoiceID := e.
		POST("/invoices").
		WithJSON(JSON{
			"price": price,
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object().
		Value("id").Number().Raw()

	return uint64(invoiceID)
}

func sendGetInvoiceRequest(t *testing.T, e *httpexpect.Expect, invoiceID uint64) *TestInvoice {
	t.Helper()

	invoicePath := fmt.Sprintf("/invoices/%d", invoiceID)

	invoice := e.
		GET(invoicePath).
		Expect().
		Status(http.StatusOK).
		JSON().Object()

	return &TestInvoice{
		t: t,

		ID:      invoice.Value("id").Number(),
		Price:   invoice.Value("price").Number(),
		Balance: invoice.Value("balance").Number(),
		Address: invoice.Value("address").String(),
		Status:  invoice.Value("status").String(),
	}
}

const transactionProcessingTime = 5 * time.Second

func waitForProcessing(t *testing.T) {
	t.Helper()

	t.Logf("let the transaction be processed for %s", transactionProcessingTime)
	time.Sleep(transactionProcessingTime)
}
