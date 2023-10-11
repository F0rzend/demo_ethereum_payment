package domain

import (
	geth "github.com/ethereum/go-ethereum/common"
	"math/big"
)

type Invoice struct {
	id      ID
	price   WEI
	balance WEI
	address Address
	status  InvoiceStatus
}

type ID = *big.Int
type WEI = *big.Int
type Address = *geth.Address

type InvoiceStatus string

const (
	InvoiceStatusPending InvoiceStatus = "pending"
	InvoiceStatusPaid    InvoiceStatus = "paid"
)

func NewInvoice(
	id ID,
	price WEI,
	balance WEI,
	address *geth.Address,
	status InvoiceStatus,
) *Invoice {
	return &Invoice{
		id:      id,
		price:   price,
		balance: balance,
		address: address,
		status:  status,
	}
}

func (i *Invoice) ID() ID {
	return i.id
}

func (i *Invoice) Price() WEI {
	return i.price
}

func (i *Invoice) Balance() WEI {
	return i.balance
}

func (i *Invoice) Address() Address {
	return i.address
}

func (i *Invoice) Status() InvoiceStatus {
	return i.status
}

func (i *Invoice) Deposit(amount WEI) {
	i.balance.Add(i.balance, amount)
	if i.balance.Cmp(i.price) >= 0 {
		i.status = InvoiceStatusPaid
	}
}
