package main

import (
	"fmt"

	"google.golang.org/api/gmail/v1"
)

func getMessageBody(payload *gmail.MessagePart) string {
	if payload == nil {
		return ""
	}
	// Prefer text/plain
	if payload.MimeType == "text/plain" && payload.Body != nil && payload.Body.Data != "" {
		return decodeBase64URL(payload.Body.Data)
	}
	// Fallback to text/html
	if payload.MimeType == "text/html" && payload.Body != nil && payload.Body.Data != "" {
		return decodeBase64URL(payload.Body.Data)
	}
	// Recursively check parts
	for _, part := range payload.Parts {
		body := getMessageBody(part)
		if body != "" {
			return body
		}
	}
	return ""
}

func processEmails(srv *gmail.Service, user string, fsClient *FirestoreClient) {
	pageToken := ""
	for {
		req := srv.Users.Messages.List(user).Q("from:alerts@hdfcbank.net OR from:customercare@icicibank.com OR from:credit_cards@icicibank.com newer_than:365d").MaxResults(500)
		if pageToken != "" {
			req = req.PageToken(pageToken)
		}
		res, err := req.Do()
		if err != nil {
			fmt.Println("Error fetching messages:", err)
			return
		}

		for _, msg := range res.Messages {
			m, _ := srv.Users.Messages.Get(user, msg.Id).Format("full").Do()

			body := getMessageBody(m.Payload)
			cleanBody := stripHTMLTags(body)

			var tx *Transaction

			if tx = parseICICICreditCardTransaction(cleanBody); tx != nil {
				fmt.Printf("✅ Parsed %s Transaction:\n%+v\n", tx.Type, *tx)
			} else if tx = parseCreditCardTransaction(cleanBody); tx != nil {
				fmt.Printf("✅ Parsed %s Transaction:\n%+v\n", tx.Type, *tx)
			} else if tx = parseCardPaymentTransaction(cleanBody); tx != nil {
				fmt.Printf("✅ Parsed %s Transaction:\n%+v\n", tx.Type, *tx)
			} else if tx = parseIMPSPaymentTransaction(cleanBody); tx != nil {
				fmt.Printf("✅ Parsed %s Transaction:\n%+v\n", tx.Type, *tx)
			} else if tx = parseBankTransaction(cleanBody); tx != nil {
				fmt.Printf("✅ Parsed %s Transaction:\n%+v\n", tx.Type, *tx)
			} else {
				fmt.Println("⚠️ No known transaction format detected.")

				// Collect headers for context (optional)
				headers := make(map[string]string)
				for _, h := range m.Payload.Headers {
					headers[h.Name] = h.Value
				}

				if err := fsClient.SaveUnparsedEmail(cleanBody, headers); err != nil {
					fmt.Println("❌ Failed to save unparsed email:", err)
				} else {
					fmt.Println("✅ Unparsed email saved to Firestore.")
				}
				continue
			}

			if err := fsClient.SaveTransaction(*tx); err != nil {
				fmt.Println("❌ Firestore save failed:", err)
			} else {
				fmt.Println("✅ Transaction saved to Firestore.")
			}
		}

		if res.NextPageToken == "" {
			break
		}
		pageToken = res.NextPageToken
	}
}
