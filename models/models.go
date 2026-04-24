package models

import (
	"time"
)

type Transaction struct {
	ID              string    `bson:"-" firestore:"-" json:"id"`
	Type            string    `bson:"type" firestore:"type" json:"type"`
	CardEnding      string    `bson:"cardending" firestore:"cardending" json:"card_ending"`
	DebitedAccount  string    `bson:"debitedaccount" firestore:"debitedaccount" json:"debited_account"`
	CreditedAccount string    `bson:"creditedaccount" firestore:"creditedaccount" json:"credited_account"`
	Amount          float64   `bson:"amount" firestore:"amount" json:"amount"`
	Vendor          string    `bson:"vendor" firestore:"vendor" json:"vendor"`
	DateTime        time.Time `bson:"datetime" firestore:"datetime" json:"date_time"`
	Category        string    `bson:"category" firestore:"category" json:"category"`
}

func (t Transaction) IsCredit() bool {
	return t.Amount < 0
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
	"zomato":                                 "Food",
	"eternal":                                "Food",
	"hdfc ergo gic other we":                 "Healthcare",
	"myntra":                                 "Shopping",
	"rameshwaram":                            "Food",
	"amazon pay in grocery":                  "Grocery",
	"amazon pay grocery":                     "Grocery",
	"amazon pay in e commerce":               "Amazon",
	"amazon pay ecom":                        "Amazon",
	"swiggy instamart":                       "Grocery",
	"swiggy":                                 "Food",
	"groww":                                  "SIP",
	"zepto":                                  "Grocery",
	"zeptonow":                               "Grocery",
	"electricity":                            "Bills",
	"sonyliv":                                "Entertainment",
	"airtel prepaid":                         "Bills",
	"airtel":                                 "Bills",
	"apple services":                         "Subscription",
	"apple media services":                   "Subscription",
	"dineout":                                "Food",
	"blinkit":                                "Grocery",
	"hopscotch":                              "Shopping",
	"licious":                                "Grocery",
	"apollo":                                 "Healthcare",
	"firstcry":                               "Shopping",
	"mtr maiyas":                             "Food",
	"travel":                                 "Travel",
	"atria convergence technologies limited": "Bills",
	"vegetables":                             "Grocery",
	"fruits":                                 "Grocery",
	"klay":                                   "School Fees",
	"mmt":                                    "Travel",
	"yatra":                                  "Travel",
	"indane gas (indian oil)":                "Bills",
	"fss4firstcry":                           "Shopping",
	"my gate":                                "Bills",
	"foods":                                  "Food",
	"food":                                   "Food",
	"sweets":                                 "Food",
	"zerodha":                                "SIP",
	"uttar pradesh power corporation limited": "Bills",
	"district dining":                         "Food",
	"stylerrio":                               "Entertainment",
	"uber":                                    "Travel",
	"rapido":                                  "Travel",
	"bar":                                     "Food",
	"noodle":                                  "Food",
	"pizza":                                   "Food",
	"rbl bank credit card":                    "Bills",
	"popeyes":                                 "Food",
	"1 mg mall":                               "Shopping",
	"adda":                                    "Bills",
	"hospital":                                "Healthcare",
	"playo":                                   "Entertainment",
	"boba":                                    "Food",
	"pvr":                                     "Entertainment",
	"fomomo":                                  "Food",
	"momo":                                    "Food",
	"district":                                "Entertainment",
	"movie":                                   "Entertainment",
	"movie ticket":                            "Entertainment",
	"gold":                                    "Shopping",
	"malabar":                                 "Shopping",
	"bookmyshow":                              "Entertainment",
	"flight":                                  "Travel",
	"irctc":                                   "Travel",
	"train":                                   "Travel",
	"bus":                                     "Travel",
	"hotels":                                  "Travel",
	"flight ticket":                           "Travel",
	"train ticket":                            "Travel",
	"bus ticket":                              "Travel",
	"hotel":                                   "Travel",
	"hotel booking":                           "Travel",
	"flight booking":                          "Travel",
	"brewpub":                                 "Food",
	"pub":                                     "Food",
	"reliance retail limited":                 "Shopping",
	"filling statio":                          "Petrol",
	"station":                                 "Petrol",
	"coco krpuram":                            "Petrol",
	"coco kr puram":                           "Petrol",
	"sai krishna service st":                  "Petrol",
	"saikrishna service sta":                  "Petrol",
	"iocl":                                    "Petrol",
}
