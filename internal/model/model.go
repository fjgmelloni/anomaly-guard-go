package model

type AnalyzeRequest struct {
	Transactions []Transaction `json:"transactions"`
}

type Transaction struct {
	ID                 string  `json:"id"`
	AccountID          string  `json:"account_id"`
	Amount             float64 `json:"amount"`
	Hour               int     `json:"hour"`
	Region             string  `json:"region"`
	KnownRegion        bool    `json:"known_region"`
	FailedAttempts10m  int     `json:"failed_attempts_10m"`
	MinutesSinceLastTx int     `json:"minutes_since_last_tx"`
}

type AnalyzeResponse struct {
	Summary Summary         `json:"summary"`
	Results []AnalyzeResult `json:"results"`
}

type Summary struct {
	Total      int `json:"total"`
	Suspicious int `json:"suspicious"`
}

type AnalyzeResult struct {
	Transaction       Transaction `json:"transaction"`
	AnomalyScore      float64     `json:"anomaly_score"`
	Suspicious        bool        `json:"suspicious"`
	RiskLevel         string      `json:"risk_level"`
	RecommendedAction string      `json:"recommended_action"`
	Priority          string      `json:"priority"`
	Reason            string      `json:"reason"`
	DecisionSource    string      `json:"decision_source"`
	DecisionError     string      `json:"decision_error,omitempty"`
}

type DecisionInput struct {
	Transaction  Transaction `json:"transaction"`
	AnomalyScore float64     `json:"anomaly_score"`
	Suspicious   bool        `json:"suspicious"`
}

type Decision struct {
	RiskLevel         string `json:"risk_level"`
	RecommendedAction string `json:"recommended_action"`
	Priority          string `json:"priority"`
	Reason            string `json:"reason"`
	Source            string `json:"source"`
	Error             string `json:"error,omitempty"`
}
