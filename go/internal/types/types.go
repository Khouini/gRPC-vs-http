package types

// User represents a user in the system
type User struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Age    int    `json:"age"`
	City   string `json:"city"`
	Active bool   `json:"active"`
}

// Metadata contains information about the dataset
type Metadata struct {
	GeneratedAt    string  `json:"generatedAt"`
	TargetSizeMB   float64 `json:"targetSizeMB"`
	EstimatedItems int     `json:"estimatedItems"`
	ActualSizeMB   float64 `json:"actualSizeMB"`
	ActualItems    int     `json:"actualItems"`
}

// DataFile represents the structure of the data.json file
type DataFile struct {
	Metadata Metadata `json:"metadata"`
	Users    []User   `json:"users"`
}

// StatsResponse represents the response from the gateway
type StatsResponse struct {
	TotalUsers    int     `json:"totalUsers"`
	ActiveUsers   int     `json:"activeUsers"`
	InactiveUsers int     `json:"inactiveUsers"`
	DataSize      float64 `json:"dataSize"`
	ProcessTimeMs int64   `json:"processTimeMs"`
}
