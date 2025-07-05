package main

type Transaction struct {
	Type            string // "CreditCard" or "BankTransfer"
	CardEnding      string
	DebitedAccount  string
	CreditedAccount string
	Amount          string
	Vendor          string
	DateTime        string
}
