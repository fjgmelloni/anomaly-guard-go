package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"anomaly-guard-go/internal/anomaly"
	"anomaly-guard-go/internal/decision"
	"anomaly-guard-go/internal/gemini"
	apihttp "anomaly-guard-go/internal/http"
)

func main() {
	port := getenv("PORT", "8080")
	addr := ":" + port

	detector := anomaly.NewDetector()
	decisionEngine := decision.NewEngine(
		gemini.NewClient(
			getenv("GEMINI_API_KEY", ""),
			getenv("GEMINI_MODEL", "gemini-2.5-flash"),
			http.DefaultClient,
		),
	)

	server := &http.Server{
		Addr:              addr,
		Handler:           apihttp.NewServer(detector, decisionEngine),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("anomaly-guard-go listening on %s", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
