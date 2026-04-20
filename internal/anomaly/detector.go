package anomaly

import (
	"math"

	"anomaly-guard-go/internal/model"
)

type Detector struct{}

func NewDetector() *Detector {
	return &Detector{}
}

func (d *Detector) Score(tx model.Transaction) float64 {
	var score float64

	score += clamp(tx.Amount/10000.0, 0, 0.35)

	if tx.Hour < 6 || tx.Hour > 22 {
		score += 0.20
	}

	if !tx.KnownRegion {
		score += 0.20
	}

	score += clamp(float64(tx.FailedAttempts10m)*0.08, 0, 0.20)

	if tx.MinutesSinceLastTx >= 0 && tx.MinutesSinceLastTx < 2 {
		score += 0.10
	}

	if tx.MinutesSinceLastTx > 720 {
		score += 0.05
	}

	return math.Round(clamp(score, 0, 0.99)*100) / 100
}

func (d *Detector) IsSuspicious(score float64) bool {
	return score >= 0.65
}

func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
