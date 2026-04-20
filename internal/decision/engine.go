package decision

import (
	"context"
	"fmt"
	"reflect"

	"anomaly-guard-go/internal/model"
)

type aiClient interface {
	Decide(ctx context.Context, input model.DecisionInput) (model.Decision, error)
}

type Engine struct {
	client aiClient
}

func NewEngine(client aiClient) *Engine {
	if isNilClient(client) {
		client = nil
	}

	return &Engine{client: client}
}

func (e *Engine) Decide(ctx context.Context, input model.DecisionInput) model.Decision {
	if input.Suspicious && e.client != nil {
		if decision, err := e.client.Decide(ctx, input); err == nil {
			decision.Source = "gemini"
			return decision
		} else {
			decision := heuristicDecision(input)
			decision.Error = fmt.Sprintf("gemini fallback: %v", err)
			return decision
		}
	}

	return heuristicDecision(input)
}

func heuristicDecision(input model.DecisionInput) model.Decision {
	tx := input.Transaction
	score := input.AnomalyScore

	decision := model.Decision{
		RiskLevel:         "low",
		RecommendedAction: "approve",
		Priority:          "P3",
		Reason:            "Transaction is within the expected operating pattern.",
		Source:            "heuristic",
	}

	switch {
	case score >= 0.90 || (tx.FailedAttempts10m >= 4 && !tx.KnownRegion):
		decision.RiskLevel = "critical"
		decision.RecommendedAction = "block"
		decision.Priority = "P1"
		decision.Reason = "High anomaly score combined with strong fraud indicators such as repeated failed attempts or an unknown region."
	case score >= 0.80:
		decision.RiskLevel = "high"
		decision.RecommendedAction = "manual_review"
		decision.Priority = "P1"
		decision.Reason = "Transaction is significantly outside the usual pattern and should be reviewed before approval."
	case score >= 0.65:
		decision.RiskLevel = "medium"
		decision.RecommendedAction = "monitor"
		decision.Priority = "P2"
		decision.Reason = "Transaction shows unusual signals but does not justify an immediate block."
	}

	return decision
}

func isNilClient(client aiClient) bool {
	if client == nil {
		return true
	}

	value := reflect.ValueOf(client)
	switch value.Kind() {
	case reflect.Ptr, reflect.Map, reflect.Slice, reflect.Interface, reflect.Func:
		return value.IsNil()
	default:
		return false
	}
}
