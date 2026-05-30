package analyzer

import (
	"context"
	"log"
	"net"
	"server/internal/database"
	"server/pb"
)

type AttackType string

const (
	AttackARPFlood     AttackType = "ARP_FLOOD"
	AttackMACSpoofing  AttackType = "MAC_SPOOFING"
	AttackARPPoisoning AttackType = "ARP_POISONING"
)

type Alert struct {
	AgentId    string
	TargetIp   string
	AttackType AttackType
	Message    string
}

type AnalysisReport struct {
	Event   *pb.ARPEvent
	Attacks map[AttackType]string
}

func (ar *AnalysisReport) RegisterAttack(attackType AttackType, finding string) {
	ar.Attacks[attackType] = finding
}

type AnalysisStep interface {
	Process(ctx context.Context, report *AnalysisReport) error
}

type Analyzer struct {
	db               database.Database
	eventChan        chan *pb.ARPEvent
	notificationChan chan *Alert
	steps            []AnalysisStep
}

func NewAnalyzer(db database.Database, eventChan chan *pb.ARPEvent, notificationChan chan *Alert) *Analyzer {
	steps := []AnalysisStep{
		&MACChangeDetectorStep{db: db},
		&MACSpoofDetectorStep{db: db},
		&FrequencyDetectorStep{
			history:   make(map[string][]uint64),
			threshold: 20,
			window:    2 * 1000, // 2s
		},
	}

	return &Analyzer{
		db:               db,
		eventChan:        eventChan,
		notificationChan: notificationChan,
		steps:            steps,
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
		Event:   event,
		Attacks: make(map[AttackType]string),
	}

	for _, step := range a.steps {
		if err := step.Process(ctx, report); err != nil {
			return err
		}
	}

	if len(report.Attacks) != 0 {
		for attack, msg := range report.Attacks {
			targetIPString := net.IP(event.TargetIp).String()

			alert := &Alert{
				AgentId:    event.AgentId,
				TargetIp:   targetIPString,
				AttackType: attack,
				Message:    msg,
			}

			a.notificationChan <- alert
		}
	}

	return nil
}
