package notifier

import (
	"context"
	"server/internal/analyzer"
	"server/internal/database"
)

type MailSender struct{}

func (sen *MailSender) SendAlert(ctx context.Context, channel database.NotificationChannel, alert *analyzer.Alert) error {
	return nil
}
