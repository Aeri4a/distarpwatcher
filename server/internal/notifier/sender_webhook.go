package notifier

import (
	"context"
	"server/internal/analyzer"
	"server/internal/database"
)

type WebhookSender struct{}

func (sen *WebhookSender) SendAlert(ctx context.Context, channel database.NotificationChannel, report *analyzer.AnalysisReport) error {
	return nil
}
