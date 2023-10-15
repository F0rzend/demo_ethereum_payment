package tests

import (
	"context"
	"testing"
	"time"
)

func (s *TestSuite) TestApplication() {
	ctx := context.Background()
	t := s.T()

	const transactionValue = 1

	invoiceID := createInvoice(t, s.e(), transactionValue*2)

	invoice := getInvoice(t, s.e(), invoiceID)
	invoice.Status.Equal(InvoiceStatusPending)

	tx := invoice.Deposit(s.eth, transactionValue)
	tx.WaitConfirmation(ctx, s.eth)
	waitForProcessing(t)

	invoice = getInvoice(t, s.e(), invoiceID)
	invoice.Status.Equal(InvoiceStatusPaid)

	tx = invoice.Deposit(s.eth, transactionValue)
	tx.WaitConfirmation(ctx, s.eth)
	waitForProcessing(t)

	invoice = getInvoice(t, s.e(), invoiceID)
	invoice.Status.Equal(InvoiceStatusPaid)
}

const transactionProcessingTime = 3 * time.Second

func waitForProcessing(t *testing.T) {
	t.Helper()

	t.Logf("let the transaction be processed for %s", transactionProcessingTime)
	time.Sleep(transactionProcessingTime)
}
