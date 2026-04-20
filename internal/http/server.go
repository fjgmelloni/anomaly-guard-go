package http

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"anomaly-guard-go/internal/model"
)

type detector interface {
	Score(tx model.Transaction) float64
	IsSuspicious(score float64) bool
}

type decisionEngine interface {
	Decide(ctx context.Context, input model.DecisionInput) model.Decision
}

func NewServer(detector detector, decisionEngine decisionEngine) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("/analyze", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}

		var request model.AnalyzeRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON payload"})
			return
		}

		if len(request.Transactions) == 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "transactions array is required"})
			return
		}

		results := make([]model.AnalyzeResult, 0, len(request.Transactions))
		suspiciousCount := 0

		for _, tx := range request.Transactions {
			score := detector.Score(tx)
			suspicious := detector.IsSuspicious(score)
			if suspicious {
				suspiciousCount++
			}

			ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
			decision := decisionEngine.Decide(ctx, model.DecisionInput{
				Transaction:  tx,
				AnomalyScore: score,
				Suspicious:   suspicious,
			})
			cancel()

			results = append(results, model.AnalyzeResult{
				Transaction:       tx,
				AnomalyScore:      score,
				Suspicious:        suspicious,
				RiskLevel:         decision.RiskLevel,
				RecommendedAction: decision.RecommendedAction,
				Priority:          decision.Priority,
				Reason:            decision.Reason,
				DecisionSource:    decision.Source,
				DecisionError:     decision.Error,
			})
		}

		response := model.AnalyzeResponse{
			Summary: model.Summary{
				Total:      len(results),
				Suspicious: suspiciousCount,
			},
			Results: results,
		}

		writeJSON(w, http.StatusOK, response)
	})

	return mux
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
