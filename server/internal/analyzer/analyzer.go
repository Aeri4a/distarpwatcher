package analyzer

import (
	"context"
	"log"
	"time"

	"server/internal/config"
	"server/internal/database"
)

type Analyzer struct {
	config config.AnalyzerConfig
	db     database.Database
}

func NewAnalyzer(cfg config.AnalyzerConfig, db database.Database) *Analyzer {
	return &Analyzer{
		config: cfg,
		db:     db,
	}
}

func (a *Analyzer) Start(ctx context.Context) error {
	log.Printf("Analyzer starting with interval: %d seconds", a.config.Interval)
	
	ticker := time.NewTicker(time.Duration(a.config.Interval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Shutting down Analyzer...")
			return ctx.Err()
		case <-ticker.C:
			if err := a.Analyze(ctx); err != nil {
				log.Printf("Analyzer error: %v", err)
				// We don't necessarily want to kill the whole server if one analysis cycle fails
			}
		}
	}
}

func (a *Analyzer) Analyze(ctx context.Context) error {
	log.Println("Running analysis cycle...")
	// TODO: Implement analysis logic
	return nil
}
