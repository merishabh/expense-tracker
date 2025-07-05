package main

import (
	"fmt"

	"google.golang.org/api/gmail/v1"
)

func getMessageBody(payload *gmail.MessagePart) string {
	if payload == nil {
		return ""
	}
	if payload.Body != nil && payload.Body.Data != "" {
		return decodeBase64URL(payload.Body.Data)
	}
	for _, part := range payload.Parts {
		body := getMessageBody(part)
		if body != "" {
			return body
		}
	}
	return ""
}

func processEmails(srv *gmail.Service, user string, fsClient *FirestoreClient) {
	req := srv.Users.Messages.List(user).Q("from:alerts@hdfcbank.net newer_than:5d")
	res, err := req.Do()
	if err != nil {
		fmt.Println("Error fetching messages:", err)
		return
	}

	for _, msg := range res.Messages {
		m, _ := srv.Users.Messages.Get(user, msg.Id).Format("full").Do()

		body := getMessageBody(m.Payload)

		fmt.Println("-----")
		for _, h := range m.Payload.Headers {
			if h.Name == "From" || h.Name == "Subject" {
				fmt.Printf("%s: %s\n", h.Name, h.Value)
			}
		}
		fmt.Println("Body:")
		// fmt.Println(body)

		if tx := parseCreditCardTransaction(body); tx != nil {
			fmt.Println("✅ Parsed Credit Card Transaction:")
			fmt.Printf("%+v\n", *tx)

			if err := fsClient.SaveTransaction(*tx); err != nil {
				fmt.Println("❌ Firestore save failed:", err)
			} else {
				fmt.Println("✅ Transaction saved to Firestore.")
			}
			continue
		}

		if tx := parseBankTransaction(body); tx != nil {
			fmt.Println("✅ Parsed Bank Transaction:")
			fmt.Printf("%+v\n", *tx)

			if err := fsClient.SaveTransaction(*tx); err != nil {
				fmt.Println("❌ Firestore save failed:", err)
			} else {
				fmt.Println("✅ Transaction saved to Firestore.")
			}
			continue
		}

		fmt.Println("⚠️ No known transaction format detected.")
	}
}
