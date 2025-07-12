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
	"flight":           "Flight",
	"airbnb":           "Stay",
	"uber":             "Transportation",
	"ola":              "Transportation",
	"rapido":           "Transportation",
	"metro":            "Transportation",
	"irctc":            "Transportation",
	"makemytrip":       "Transportation",
	"goibibo":          "Transportation",
	"cleartrip":        "Transportation",
	"redbus":           "Transportation",
	"petrol pump":      "Transportation",
	"shell":            "Transportation",
	"hp":               "Transportation",
	"indian oil":       "Transportation",
	"bharat petroleum": "Transportation",

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
	"electricity": "Utilities",
	"water":       "Utilities",
	"gas":         "Utilities",
	"broadband":   "Utilities",
	"jio":         "Utilities",
	"airtel":      "Utilities",
	"vodafone":    "Utilities",
	"bsnl":        "Utilities",
	"wifi":        "Utilities",

	// Healthcare
	"apollo":    "Healthcare",
	"fortis":    "Healthcare",
	"max":       "Healthcare",
	"manipal":   "Healthcare",
	"pharmeasy": "Healthcare",
	"netmeds":   "Healthcare",
	"1mg":       "Healthcare",
	"medplus":   "Healthcare",

	// Education
	"byju":       "Education",
	"unacademy":  "Education",
	"vedantu":    "Education",
	"coursera":   "Education",
	"udemy":      "Education",
	"skillshare": "Education",

	// Finance
	"sip":              "Investment",
	"mutual fund":      "Investment",
	"fd":               "Investment",
	"insurance":        "Insurance",
	"lic":              "Insurance",
	"hdfc life":        "Insurance",
	"icici prudential": "Insurance",
	"bescom":           "Electricity",
}

// CategorizeTransaction determines the category based on vendor name
func CategorizeTransaction(vendor string) string {
	if vendor == "" {
		return "Other"
	}

	// Convert to lowercase for case-insensitive matching
	vendorLower := strings.ToLower(vendor)

	// Check for exact matches first
	if category, exists := VendorCategoryMapping[vendorLower]; exists {
		return category
	}

	// Check for partial matches
	for mappedVendor, category := range VendorCategoryMapping {
		if strings.Contains(vendorLower, mappedVendor) || strings.Contains(mappedVendor, vendorLower) {
			return category
		}
	}

	return "Other"
}
