package analyzer

import (
	"context"
	"log"
	"server/internal/database"
	"server/pb"
)

type AnalysisReport struct {
	Event    *pb.ARPEvent
	Flags    uint32 // skipped for now
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
			//log.Printf("Analyzer received event %v", event)
			if err := a.AnalyzeCycle(ctx, event); err != nil {
				log.Printf("Analyzer error: %v", err)
			}
		}
	}
}

func (a *Analyzer) AnalyzeCycle(ctx context.Context, event *pb.ARPEvent) error {
	report := &AnalysisReport{
		Event:    event,
		Flags:    0,
		Findings: []string{},
	}

	for _, step := range a.steps {
		if err := step.Process(ctx, report); err != nil {
			return err
		}
	}

	if len(report.Findings) != 0 {
		log.Printf("Found findings: %v", report.Findings)
	}

	return nil
}
