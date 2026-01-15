package services

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/yourusername/expense-tracker/ai"
	"github.com/yourusername/expense-tracker/models"
)

func ParseCreditCardTransaction(text string, dbClient models.DatabaseClient, geminiClient *ai.GeminiClient) *models.Transaction {
	// Try new HDFC format: "Rs.304.00 is debited from your HDFC Bank Credit Card ending 4207 towards RAZORPAY LICIOUS on 09 Jan, 2026 at 16:28:26."
	re := regexp.MustCompile(`Rs\.?([\d,\.]+)\s+is\s+debited\s+from\s+your\s+HDFC\s+Bank\s+Credit\s+Card\s+ending\s+(\d+)\s+towards\s+(.+?)\s+on\s+(\d{1,2}\s+[A-Za-z]{3},\s+\d{4})\s+at\s+(\d{2}:\d{2}:\d{2})`)
	match := re.FindStringSubmatch(text)
	if len(match) == 6 {
		amount, err := strconv.ParseFloat(strings.ReplaceAll(match[1], ",", ""), 64)
		if err != nil {
			log.Printf("Error parsing amount: %v", err)
			return nil
		}
		// Parse date format: "09 Jan, 2026 at 16:28:26"
		datetimeStr := match[4] + " " + match[5]
		dt, err := time.Parse("2 Jan, 2006 15:04:05", datetimeStr)
		if err != nil {
			log.Printf("Error parsing date: %v", err)
			return nil
		}
		vendor := strings.TrimSpace(match[3])
		return &models.Transaction{
			Type:       "HDFCCreditCard",
			CardEnding: match[2],
			Amount:     amount,
			Vendor:     vendor,
			DateTime:   dt,
			Category:   CategorizeTransaction(vendor, dbClient, geminiClient),
		}
	}

	// Try original format: "Credit Card ending 1234 for Rs 100.00 at VENDOR on 01-01-2024 12:00:00"
	re = regexp.MustCompile(`Credit Card ending (\d+) for Rs ([\d,.]+) at (.*?) on (\d{2}-\d{2}-\d{4} \d{2}:\d{2}:\d{2})`)
	match = re.FindStringSubmatch(text)
	if len(match) == 5 {
		amount, err := strconv.ParseFloat(strings.ReplaceAll(match[2], ",", ""), 64)
		if err != nil {
			log.Printf("Error parsing amount: %v", err)
			return nil
		}
		dt, err := time.Parse("02-01-2006 15:04:05", match[4])
		if err != nil {
			log.Printf("Error parsing date: %v", err)
			return nil
		}
		vendor := strings.TrimSpace(match[3])
		return &models.Transaction{
			Type:       "HDFCCreditCard",
			CardEnding: match[1],
			Amount:     amount,
			Vendor:     vendor,
			DateTime:   dt,
			Category:   CategorizeTransaction(vendor, dbClient, geminiClient),
		}
	}
	return nil
}

func ParseBankTransaction(text string, dbClient models.DatabaseClient, geminiClient *ai.GeminiClient) *models.Transaction {
	re := regexp.MustCompile(`Your A/c (\w+) is debited for INR ([\d,\.]+) on (\d{2}-\d{2}-\d{2}) and A/c (\w+) is credited`)
	match := re.FindStringSubmatch(text)
	if len(match) == 5 {
		amount, err := strconv.ParseFloat(strings.ReplaceAll(match[2], ",", ""), 64)
		if err != nil {
			log.Printf("Error parsing amount: %v", err)
			return nil
		}
		// Parse date (assuming format is DD-MM-YY)
		dateStr := match[3]
		dt, err := time.Parse("02-01-06", dateStr)
		if err != nil {
			log.Printf("Error parsing date: %v", err)
			return nil
		}
		return &models.Transaction{
			Type:            "BankTransfer",
			DebitedAccount:  match[1],
			CreditedAccount: match[4],
			Amount:          amount,
			DateTime:        dt,
		}
	}
	return nil
}

func ParseICICICreditCardTransaction(text string, dbClient models.DatabaseClient, geminiClient *ai.GeminiClient) *models.Transaction {
	// Updated regex to stop vendor capture at first period followed by " The" (before "The Available Credit Limit")
	re := regexp.MustCompile(`ICICI Bank Credit Card (\w+) has been used for a transaction of INR ([\d,\.]+) on ([A-Za-z]+ \d{1,2}, \d{4}) at (\d{2}:\d{2}:\d{2})\. Info: (.+?)\.\s+The`)
	match := re.FindStringSubmatch(text)
	if len(match) == 6 {
		amount, err := strconv.ParseFloat(strings.ReplaceAll(match[2], ",", ""), 64)
		if err != nil {
			log.Printf("Error parsing amount: %v", err)
			return nil
		}
		// Parse date and time
		datetimeStr := match[3] + " " + match[4]
		dt, err := time.Parse("Jan 2, 2006 15:04:05", datetimeStr)
		if err != nil {
			log.Printf("Error parsing date: %v", err)
			return nil
		}
		vendor := strings.TrimSpace(match[5])
		return &models.Transaction{
			Type:       "ICICICreditCard",
			CardEnding: match[1],
			Amount:     amount,
			Vendor:     vendor,
			DateTime:   dt,
			Category:   CategorizeTransaction(vendor, dbClient, geminiClient),
		}
	}
	return nil
}

// Card Payment Transaction
func ParseCardPaymentTransaction(text string, dbClient models.DatabaseClient, geminiClient *ai.GeminiClient) *models.Transaction {
	re := regexp.MustCompile(`payment of [₹INR ]*([\d,\.]+) using iMobile towards (\w+) from your Account (\w+)`)
	match := re.FindStringSubmatch(text)
	if len(match) == 4 {
		amount, err := strconv.ParseFloat(strings.ReplaceAll(match[1], ",", ""), 64)
		if err != nil {
			log.Printf("Error parsing amount: %v", err)
			return nil
		}
		vendor := match[2]
		return &models.Transaction{
			Type:           "ICICIBankTransfer",
			Amount:         amount,
			CardEnding:     vendor,
			DebitedAccount: match[3],
			Vendor:         vendor,
			Category:       CategorizeTransaction(vendor, dbClient, geminiClient),
		}
	}
	return nil
}

// IMPS Payment Transaction
func ParseIMPSPaymentTransaction(text string, dbClient models.DatabaseClient, geminiClient *ai.GeminiClient) *models.Transaction {
	re := regexp.MustCompile(`You have made an online IMPS payment of Rs ([\d,\.]+) towards (.+) on ([A-Za-z]+ \d{2}, \d{4}) at (\d{2}:\d{2}) (a\.m\.|p\.m\.) from your .* Account (\w+)`)
	match := re.FindStringSubmatch(text)
	if len(match) == 7 {
		amount, err := strconv.ParseFloat(strings.ReplaceAll(match[1], ",", ""), 64)
		if err != nil {
			log.Printf("Error parsing amount: %v", err)
			return nil
		}
		timeStr := match[4]
		if match[5] == "p.m." && !strings.HasPrefix(timeStr, "12") {
			hour, min := timeStr[:2], timeStr[3:]
			hourInt, _ := strconv.Atoi(hour)
			hourInt += 12
			timeStr = fmt.Sprintf("%02d:%s", hourInt, min)
		}
		datetimeStr := match[3] + " " + timeStr
		dt, err := time.Parse("Jan 2, 2006 15:04", datetimeStr)
		if err != nil {
			log.Printf("Error parsing date: %v", err)
			return nil
		}
		vendor := match[2] // payee
		return &models.Transaction{
			Type:           "ICICIIMPS",
			Amount:         amount,
			Vendor:         vendor,
			DateTime:       dt,
			DebitedAccount: match[6],
			Category:       CategorizeTransaction(vendor, dbClient, geminiClient),
		}
	}
	return nil
}

// CategorizeTransaction determines the category based on vendor name
// Uses manual mapping first, then MongoDB cache, then Gemini AI as fallback
func CategorizeTransaction(vendor string, dbClient models.DatabaseClient, geminiClient *ai.GeminiClient) string {
	if vendor == "" {
		return "Other"
	}

	// Convert to lowercase for case-insensitive matching
	vendorLower := strings.ToLower(vendor)

	// 1. Check manual mapping first
	if category, exists := models.VendorCategoryMapping[vendorLower]; exists {
		return category
	}

	// Check for partial matches in manual mapping
	for mappedVendor, category := range models.VendorCategoryMapping {
		if strings.Contains(vendorLower, mappedVendor) || strings.Contains(mappedVendor, vendorLower) {
			return category
		}
	}

	// 2. Check MongoDB cache
	if dbClient != nil {
		if cachedMapping, err := dbClient.GetCategoryMapping(vendorLower); err == nil && cachedMapping != nil {
			return cachedMapping.Category
		}
	}

	// 3. Use Gemini AI as fallback
	if geminiClient != nil {
		if aiCategory, err := geminiClient.ClassifyVendor(vendor); err == nil && aiCategory != "" {
			// Save AI-generated mapping to MongoDB cache
			if dbClient != nil {
				mapping := &models.CategoryMapping{
					Vendor:   vendorLower,
					Category: aiCategory,
					Source:   "ai",
					Created:  time.Now(),
				}
				dbClient.SaveCategoryMapping(mapping)
			}
			return aiCategory
		}
	}

	return "Other"
}
