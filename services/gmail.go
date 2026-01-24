package services

import (
	"fmt"

	"github.com/yourusername/expense-tracker/ai"
	"github.com/yourusername/expense-tracker/models"
	"github.com/yourusername/expense-tracker/utils"

	"google.golang.org/api/gmail/v1"
)

func getMessageBody(payload *gmail.MessagePart) string {
	if payload == nil {
		return ""
	}
	// Prefer text/plain
	if payload.MimeType == "text/plain" && payload.Body != nil && payload.Body.Data != "" {
		return utils.DecodeBase64URL(payload.Body.Data)
	}
	// Fallback to text/html
	if payload.MimeType == "text/html" && payload.Body != nil && payload.Body.Data != "" {
		return utils.DecodeBase64URL(payload.Body.Data)
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

func ProcessEmails(srv *gmail.Service, user string, dbClient models.DatabaseClient, groqClient *ai.GroqClient) {
	pageToken := ""
	for {
		req := srv.Users.Messages.List(user).Q("from:alerts@hdfcbank.net OR from:customercare@icicibank.com OR from:credit_cards@icicibank.com newer_than:100d").MaxResults(500)
		if pageToken != "" {
			req = req.PageToken(pageToken)
		}
		res, err := req.Do()
		if err != nil {
			fmt.Println("Error fetching messages:", err)
			return
		}

		for _, msg := range res.Messages {
			m, err := srv.Users.Messages.Get(user, msg.Id).Format("full").Do()
			if err != nil {
				fmt.Printf("Error fetching message %s: %v\n", msg.Id, err)
				continue
			}

			body := getMessageBody(m.Payload)
			if body == "" {
				fmt.Printf("⚠️ Empty body for message %s, skipping\n", msg.Id)
				continue
			}
			cleanBody := utils.StripHTMLTags(body)

			var tx *models.Transaction

			if tx = ParseICICICreditCardTransaction(cleanBody, dbClient, groqClient); tx != nil {
				fmt.Printf("✅ Parsed %s Transaction:\n%+v\n", tx.Type, *tx)
			} else if tx = ParseCreditCardTransaction(cleanBody, dbClient, groqClient); tx != nil {
				fmt.Printf("✅ Parsed %s Transaction:\n%+v\n", tx.Type, *tx)
			} else if tx = ParseCardPaymentTransaction(cleanBody, dbClient, groqClient); tx != nil {
				fmt.Printf("✅ Parsed %s Transaction:\n%+v\n", tx.Type, *tx)
			} else if tx = ParseIMPSPaymentTransaction(cleanBody, dbClient, groqClient); tx != nil {
				fmt.Printf("✅ Parsed %s Transaction:\n%+v\n", tx.Type, *tx)
			} else if tx = ParseBankTransaction(cleanBody, dbClient, groqClient); tx != nil {
				fmt.Printf("✅ Parsed %s Transaction:\n%+v\n", tx.Type, *tx)
			} else {
				fmt.Println("⚠️ No known transaction format detected.")

				// Collect headers for context (optional)
				headers := make(map[string]string)
				for _, h := range m.Payload.Headers {
					headers[h.Name] = h.Value
				}

				if err := dbClient.SaveUnparsedEmail(cleanBody, headers); err != nil {
					fmt.Println("Failed to save unparsed email:", err)
				} else {
					fmt.Println("Unparsed email saved to database.")
				}
				continue
			}

			if err := dbClient.SaveTransaction(*tx); err != nil {
				fmt.Println("Database save failed:", err)
			} else {
				fmt.Println("Transaction saved to database.")
			}
		}
		if res.NextPageToken == "" {
			break
		}
		pageToken = res.NextPageToken
	}
}
