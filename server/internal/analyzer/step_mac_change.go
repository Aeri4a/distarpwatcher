package analyzer

import (
	"context"
	"server/internal/database"
)

type MACChangeDetectorStep struct {
	db database.Database
}

func (macdet *MACChangeDetectorStep) Process(ctx context.Context, report *AnalysisReport) error {
	return nil
}
