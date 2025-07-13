package main

import (
	"strings"
	"time"
)

type Transaction struct {
	Type            string // "CreditCard" or "BankTransfer"
	CardEnding      string
	DebitedAccount  string
	CreditedAccount string
	Amount          float64
	Vendor          string
	DateTime        time.Time
	Category        string // New field for transaction category
}

// CategoryMapping represents a vendor-to-category mapping stored in MongoDB
type CategoryMapping struct {
	Vendor   string    `bson:"vendor" json:"vendor"`
	Category string    `bson:"category" json:"category"`
	Source   string    `bson:"source" json:"source"` // "manual" or "ai"
	Created  time.Time `bson:"created" json:"created"`
}

// VendorCategoryMapping maps vendor names to categories
var VendorCategoryMapping = map[string]string{
	// Food & Dining
	"zomato":          "Food",
	"swiggy":          "Food",
	"dominos":         "Food",
	"mcdonalds":       "Food",
	"kfc":             "Food",
	"subway":          "Food",
	"pizza hut":       "Food",
	"burger king":     "Food",
	"dunkin":          "Food",
	"starbucks":       "Food",
	"cafe coffee day": "Food",
	"barbeque nation": "Food",
	"haldirams":       "Food",
	"blinkit":         "General_food",
	"zepto":           "General_food",
	"dineout":         "Food",
	"licious":         "Food",

	// Transportation
	"flight":           "Travel",
	"airbnb":           "Travel",
	"uber":             "Travel",
	"ola":              "Travel",
	"rapido":           "Travel",
	"metro":            "Travel",
	"irctc":            "Travel",
	"makemytrip":       "Travel",
	"goibibo":          "Travel",
	"cleartrip":        "Travel",
	"redbus":           "Travel",
	"petrol pump":      "Travel",
	"shell":            "Travel",
	"hp":               "Travel",
	"indian oil":       "Travel",
	"bharat petroleum": "Travel",

	// Shopping
	"amazon":     "Amazon",
	"flipkart":   "Shopping",
	"myntra":     "Shopping",
	"ajio":       "Shopping",
	"nykaa":      "Shopping",
	"reliance":   "Shopping",
	"big bazaar": "Shopping",
	"dmart":      "Shopping",
	"more":       "Shopping",
	"lifestyle":  "Shopping",
	"pantaloons": "Shopping",
	"westside":   "Shopping",

	// Entertainment
	"netflix":        "Entertainment",
	"amazon prime":   "Entertainment",
	"disney hotstar": "Entertainment",
	"sony liv":       "Entertainment",
	"zee5":           "Entertainment",
	"voot":           "Entertainment",
	"bookmyshow":     "Entertainment",
	"paytm movies":   "Entertainment",
	"pvr":            "Entertainment",
	"inox":           "Entertainment",

	// Utilities
	"electricity": "Bills",
	"water":       "Bills",
	"gas":         "Bills",
	"broadband":   "Bills",
	"jio":         "Bills",
	"airtel":      "Bills",
	"vodafone":    "Bills",
	"bsnl":        "Bills",
	"wifi":        "Bills",

	// Healthcare
	"apollo":    "Healthcare",
	"fortis":    "Healthcare",
	"max":       "Healthcare",
	"manipal":   "Healthcare",
	"pharmeasy": "Healthcare",
	"netmeds":   "Healthcare",
	"1mg":       "Healthcare",
	"medplus":   "Healthcare",

	// Finance
	"sip":              "Other",
	"mutual fund":      "Other",
	"fd":               "Other",
	"insurance":        "Bills",
	"lic":              "Bills",
	"hdfc life":        "Bills",
	"icici prudential": "Bills",
	"bescom":           "Bills",
}

// CategorizeTransaction determines the category based on vendor name
// Uses manual mapping first, then MongoDB cache, then Gemini AI as fallback
func CategorizeTransaction(vendor string, dbClient DatabaseClient, geminiClient *GeminiClient) string {
	if vendor == "" {
		return "Other"
	}

	// Convert to lowercase for case-insensitive matching
	vendorLower := strings.ToLower(vendor)

	// 1. Check manual mapping first
	if category, exists := VendorCategoryMapping[vendorLower]; exists {
		return category
	}

	// Check for partial matches in manual mapping
	for mappedVendor, category := range VendorCategoryMapping {
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
				mapping := &CategoryMapping{
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
