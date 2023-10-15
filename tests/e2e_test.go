package tests

import (
	"context"
	"testing"
	"time"
)

func (s *TestSuite) TestApplication() {
	ctx := context.Background()

	const transactionValue = 1

	invoiceID := createInvoice(s.T(), s.e(), transactionValue*2)

	invoice := getInvoice(s.T(), s.e(), invoiceID)
	invoice.Status.Equal(InvoiceStatusPending)

	tx := invoice.Deposit(s.eth, transactionValue)
	tx.WaitConfirmation(ctx, s.eth)
	waitForProcessing(s.T())

	invoice = getInvoice(s.T(), s.e(), invoiceID)
	invoice.Status.Equal(InvoiceStatusPending)

	tx = invoice.Deposit(s.eth, transactionValue)
	tx.WaitConfirmation(ctx, s.eth)
	waitForProcessing(s.T())

	invoice = getInvoice(s.T(), s.e(), invoiceID)
	invoice.Status.Equal(InvoiceStatusPaid)
}

const transactionProcessingTime = 5 * time.Second

func waitForProcessing(t *testing.T) {
	t.Helper()

	t.Logf("let the transaction be processed for %s", transactionProcessingTime)
	time.Sleep(transactionProcessingTime)
}
