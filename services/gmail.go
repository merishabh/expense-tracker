package services

import (
	"fmt"
	"log"

	"github.com/yourusername/expense-tracker/models"
	"github.com/yourusername/expense-tracker/utils"

	"google.golang.org/api/gmail/v1"
)

type EmailSyncStats struct {
	EmailsFetched      int `json:"emails_fetched"`
	TransactionsParsed int `json:"transactions_parsed"`
	TransactionsSaved  int `json:"transactions_saved"`
	ParseFailures      int `json:"parse_failures"`
	SaveFailures       int `json:"save_failures"`
}

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

func ProcessEmails(srv *gmail.Service, user string, dbClient models.DatabaseClient) (EmailSyncStats, error) {
	stats := EmailSyncStats{}
	pageToken := ""
	for {
		req := srv.Users.Messages.List(user).Q("(from:alerts@hdfcbank.net OR from:alerts@hdfcbank.bank.in) newer_than:60d").MaxResults(500)
		if pageToken != "" {
			req = req.PageToken(pageToken)
		}
		res, err := req.Do()
		if err != nil {

			return stats, fmt.Errorf("error fetching gmail messages: %w", err)
		}

		for _, msg := range res.Messages {
			stats.EmailsFetched++
			m, err := srv.Users.Messages.Get(user, msg.Id).Format("full").Do()
			if err != nil {
				fmt.Printf("Error fetching message %s: %v\n", msg.Id, err)
				stats.ParseFailures++
				continue
			}

			body := getMessageBody(m.Payload)
			if body == "" {
				fmt.Printf("⚠️ Empty body for message %s, skipping\n", msg.Id)
				stats.ParseFailures++
				continue
			}
			cleanBody := utils.StripHTMLTags(body)

			var tx *models.Transaction

			if tx = ParseCreditCardTransaction(cleanBody, dbClient); tx != nil {
				fmt.Printf("✅ Parsed %s Transaction:\n%+v\n", tx.Type, *tx)
			} else if tx = ParseBankTransaction(cleanBody, dbClient); tx != nil {
				fmt.Printf("✅ Parsed %s Transaction:\n%+v\n", tx.Type, *tx)
			} else {
				fmt.Println("⚠️ No known transaction format detected.")
				stats.ParseFailures++

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
			stats.TransactionsParsed++

			if err := dbClient.SaveTransaction(*tx); err != nil {
				fmt.Println("Database save failed:", err)
				stats.SaveFailures++
			} else {
				fmt.Println("Transaction saved to database.")
				stats.TransactionsSaved++
			}
		}
		if res.NextPageToken == "" {
			break
		}
		pageToken = res.NextPageToken
	}

	log.Printf("gmail sync summary: fetched=%d parsed=%d saved=%d parse_failures=%d save_failures=%d",
		stats.EmailsFetched,
		stats.TransactionsParsed,
		stats.TransactionsSaved,
		stats.ParseFailures,
		stats.SaveFailures,
	)

	return stats, nil
}
