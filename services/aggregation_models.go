package services

// SpendResult represents total spending for a period
type SpendResult struct {
	Period     string  `json:"period"`
	TotalSpent float64 `json:"total_spent"`
}

// CategorySpendResult represents spending for a specific category
type CategorySpendResult struct {
	Category   string  `json:"category"`
	Period     string  `json:"period"`
	TotalSpent float64 `json:"total_spent"`
	Average    float64 `json:"average"`
}

// ComparisonResult represents a comparison between two periods
type ComparisonResult struct {
	BasePeriod    string  `json:"base_period"`
	ComparePeriod string  `json:"compare_period"`
	BaseAmount    float64 `json:"base_amount"`
	CompareAmount float64 `json:"compare_amount"`
	DeltaPercent  float64 `json:"delta_percent"`
}

// TopMerchantsResult represents top merchants by spending
type TopMerchantsResult struct {
	Period    string             `json:"period"`
	Merchants map[string]float64 `json:"merchants"`
}

// VendorSpendResult represents spending for a specific vendor
type VendorSpendResult struct {
	Vendor     string  `json:"vendor"`
	Period     string  `json:"period"`
	TotalSpent float64 `json:"total_spent"`
	Count      int     `json:"count"`
	Average    float64 `json:"average"`
}
