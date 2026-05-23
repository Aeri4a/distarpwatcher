package analyzer

import (
	"context"
	"log"
	"server/internal/database"
	"server/pb"
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
	db        database.Database
	eventChan chan *pb.ARPEvent
	steps     []AnalysisStep
}

func NewAnalyzer(db database.Database, eventChan chan *pb.ARPEvent) *Analyzer {
	steps := []AnalysisStep{
		&MACChangeDetectorStep{db: db},
	}

	return &Analyzer{
		db:        db,
		eventChan: eventChan,
		steps:     steps,
	}
}

func (a *Analyzer) Start(ctx context.Context) error {
	log.Println("Analyzer started. Waiting for events...")

	for {
		select {
		case <-ctx.Done():
			log.Println("Shutting down Analyzer...")
			return ctx.Err()
		case event, ok := <-a.eventChan:
			if !ok {
				log.Println("Analyzer channel closed.")
				return nil
			}
			log.Printf("Analyzer received event %v", event)
			//if err := a.AnalyzeCycle(ctx); err != nil {
			//	log.Printf("Analyzer error: %v", err)
			//}
		}
	}
}

func (a *Analyzer) AnalyzeCycle(ctx context.Context) error {
	log.Println("Running analysis cycle...")
	// report
	// step process
	// handle results
	return nil
}
