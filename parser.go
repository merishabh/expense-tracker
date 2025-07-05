package main

import (
	"regexp"
)

func parseCreditCardTransaction(text string) *Transaction {
	re := regexp.MustCompile(`Credit Card ending (\d+).*for (Rs\s[\d,.]+).*at (.*?) on (\d{2}-\d{2}-\d{4} \d{2}:\d{2}:\d{2})`)
	match := re.FindStringSubmatch(text)
	if len(match) == 5 {
		return &Transaction{
			Type:       "CreditCard",
			CardEnding: match[1],
			Amount:     match[2],
			Vendor:     match[3],
			DateTime:   match[4],
		}
	}
	return nil
}

func parseBankTransaction(text string) *Transaction {
	re := regexp.MustCompile(`A/c (\w+) is debited for (INR\s[\d,\.]+) on (\d{2}-\d{2}-\d{2}).*A/c (\w+) is credited`)
	match := re.FindStringSubmatch(text)
	if len(match) == 5 {
		return &Transaction{
			Type:            "BankTransfer",
			DebitedAccount:  match[1],
			Amount:          match[2],
			DateTime:        match[3],
			CreditedAccount: match[4],
		}
	}
	return nil
}
