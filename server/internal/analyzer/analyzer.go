package analyzer

import (
	"context"
	"log"
	"server/pb"
	"time"

	"server/internal/config"
	"server/internal/database"
)

type AnalysisReport struct {
	Event    *pb.ARPEvent
	Flags    uint32
	Findings []string
}

func (ar *AnalysisReport) AddFinding(finding string) {
	ar.Findings = append(ar.Findings, finding)
}

type AnalysisStep interface {
	Process(ctx context.Context, report *AnalysisReport) error
}

type Analyzer struct {
	config config.AnalyzerConfig
	db     database.Database
	steps  []AnalysisStep
}

func NewAnalyzer(cfg config.AnalyzerConfig, db database.Database) *Analyzer {
	steps := []AnalysisStep{
		&MACChangeDetectorStep{db: db},
	}

	return &Analyzer{
		config: cfg,
		db:     db,
		steps:  steps,
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
			if err := a.AnalyzeCycle(ctx); err != nil {
				log.Printf("Analyzer error: %v", err)
			}
		}
	}
}

func (a *Analyzer) AnalyzeCycle(ctx context.Context) error {
	log.Println("Running analysis cycle...")
	// report
	// step process
	// andle results
	return nil
}
