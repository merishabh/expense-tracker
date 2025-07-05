package main

import "time"

type Transaction struct {
	Type            string // "CreditCard" or "BankTransfer"
	CardEnding      string
	DebitedAccount  string
	CreditedAccount string
	Amount          float64
	Vendor          string
	DateTime        time.Time
}
