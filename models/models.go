package models

import (
	"time"
)

type Transaction struct {
	Type            string    `bson:"type" json:"type"` // "CreditCard" or "BankTransfer"
	CardEnding      string    `bson:"cardending" json:"card_ending"`
	DebitedAccount  string    `bson:"debitedaccount" json:"debited_account"`
	CreditedAccount string    `bson:"creditedaccount" json:"credited_account"`
	Amount          float64   `bson:"amount" json:"amount"`
	Vendor          string    `bson:"vendor" json:"vendor"`
	DateTime        time.Time `bson:"datetime" json:"date_time"`
	Category        string    `bson:"category" json:"category"` // New field for transaction category
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
	"zomato":          "Ordered_Food",
	"swiggy":          "Ordered_Food",
	"dominos":         "Ordered_Food",
	"mcdonalds":       "Ordered_Food",
	"kfc":             "Ordered_Food",
	"subway":          "Ordered_Food",
	"pizza hut":       "Ordered_Food",
	"burger king":     "Ordered_Food",
	"dunkin":          "Ordered_Food",
	"starbucks":       "Ordered_Food",
	"cafe coffee day": "Ordered_Food",
	"barbeque nation": "Ordered_Food",
	"haldirams":       "Ordered_Food",
	"eternal limited": "Ordered_Food",
	"ETERNAL LIMITED": "Ordered_Food",
	"blinkit":         "Grocery",
	"zepto":           "Grocery",
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
