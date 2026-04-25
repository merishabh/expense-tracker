package services

import (
	"fmt"
	"log"
	"strings"

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

func getHeaderValue(headers []*gmail.MessagePartHeader, name string) string {
	for _, h := range headers {
		if strings.EqualFold(h.Name, name) {
			return h.Value
		}
	}
	return ""
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
	query := "((from:alerts@hdfcbank.net OR from:alerts@hdfcbank.bank.in) OR (from:credit_cards@icicibank.com OR from:credit_cards@icici.bank.in) OR (from:RBLAlerts@rbl.bank.in)) newer_than:1d"

	log.Printf("gmail sync started user=%s query=%q", user, query)
	for {
		req := srv.Users.Messages.List(user).Q(query).MaxResults(500)
		if pageToken != "" {
			req = req.PageToken(pageToken)
		}
		res, err := req.Do()
		if err != nil {
			log.Printf("gmail sync list failed user=%s page_token_present=%t err=%v", user, pageToken != "", err)
			return stats, fmt.Errorf("error fetching gmail messages: %w", err)
		}

		log.Printf("gmail sync page fetched user=%s messages=%d next_page_token_present=%t", user, len(res.Messages), res.NextPageToken != "")

		for _, msg := range res.Messages {
			stats.EmailsFetched++
			m, err := srv.Users.Messages.Get(user, msg.Id).Format("full").Do()
			if err != nil {
				log.Printf("gmail sync message fetch failed message_id=%s err=%v", msg.Id, err)
				stats.ParseFailures++
				continue
			}

			subject := getHeaderValue(m.Payload.Headers, "Subject")
			from := getHeaderValue(m.Payload.Headers, "From")

			body := getMessageBody(m.Payload)
			if body == "" {
				log.Printf("gmail sync empty body message_id=%s from=%q subject=%q", msg.Id, from, subject)
				stats.ParseFailures++
				continue
			}
			cleanBody := utils.StripHTMLTags(body)

			var tx *models.Transaction

			if tx = ParseCreditCardTransaction(cleanBody, dbClient); tx != nil {
				log.Printf("gmail sync parsed message_id=%s from=%q subject=%q type=%s vendor=%q amount=%.2f", msg.Id, from, subject, tx.Type, tx.Vendor, tx.Amount)
			} else if tx = ParseICICICreditCardTransaction(cleanBody, dbClient); tx != nil {
				log.Printf("gmail sync parsed message_id=%s from=%q subject=%q type=%s vendor=%q amount=%.2f", msg.Id, from, subject, tx.Type, tx.Vendor, tx.Amount)
			} else if tx = ParseRBLCreditCardTransaction(cleanBody, dbClient); tx != nil {
				log.Printf("gmail sync parsed message_id=%s from=%q subject=%q type=%s vendor=%q amount=%.2f", msg.Id, from, subject, tx.Type, tx.Vendor, tx.Amount)
			} else if tx = ParseBankTransaction(cleanBody, dbClient); tx != nil {
				log.Printf("gmail sync parsed message_id=%s from=%q subject=%q type=%s vendor=%q amount=%.2f", msg.Id, from, subject, tx.Type, tx.Vendor, tx.Amount)
			} else {
				log.Printf("gmail sync unparsed message_id=%s from=%q subject=%q", msg.Id, from, subject)
				stats.ParseFailures++

				headers := make(map[string]string)
				for _, h := range m.Payload.Headers {
					headers[h.Name] = h.Value
				}

				if err := dbClient.SaveUnparsedEmail(cleanBody, headers); err != nil {
					log.Printf("gmail sync save unparsed email failed message_id=%s err=%v", msg.Id, err)
				} else {
					log.Printf("gmail sync saved unparsed email message_id=%s", msg.Id)
				}
				continue
			}
			stats.TransactionsParsed++

			if err := dbClient.SaveTransaction(*tx); err != nil {
				log.Printf("gmail sync transaction save failed message_id=%s type=%s vendor=%q amount=%.2f err=%v", msg.Id, tx.Type, tx.Vendor, tx.Amount, err)
				stats.SaveFailures++
			} else {
				log.Printf("gmail sync transaction saved message_id=%s type=%s vendor=%q amount=%.2f", msg.Id, tx.Type, tx.Vendor, tx.Amount)
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
