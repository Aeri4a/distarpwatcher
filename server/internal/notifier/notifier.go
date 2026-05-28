package notifier

import (
	"context"
	"log"
	"server/internal/analyzer"
	"server/internal/database"
)

type Sender interface {
	SendAlert(ctx context.Context, channel database.NotificationChannel, report *analyzer.AnalysisReport) error
}

type Notifier struct {
	db               database.Database
	notificationChan chan *analyzer.AnalysisReport

	senders map[string]Sender
}

func NewNotifier(db database.Database, notificationChan chan *analyzer.AnalysisReport) *Notifier {
	notif := &Notifier{
		db:               db,
		notificationChan: notificationChan,
		senders:          make(map[string]Sender),
	}

	notif.senders["MAIL"] = &MailSender{}
	notif.senders["WEBHOOK"] = &WebhookSender{}

	return notif
}

func (notif *Notifier) Start(ctx context.Context) error {
	log.Println("Notifier started. Waiting for alerts...")

	for {
		select {
		case <-ctx.Done():
			log.Println("Shutting down Notifier...")
			return ctx.Err()
		case report, ok := <-notif.notificationChan:
			if !ok {
				log.Println("Notification channel closed.")
				return nil
			}

			channels, err := notif.db.GetActiveChannels(ctx)
			if err != nil {
				log.Println("Error getting active channels: ", err)
				continue
			}

			if len(channels) == 0 {
				log.Println("No active channels found. Won't send alert.")
				continue
			}

			for _, channel := range channels {
				sender, exist := notif.senders[channel.Type]
				if !exist {
					log.Println("Sender not found: ", channel.Type)
					continue
				}

				go func(ch database.NotificationChannel, r *analyzer.AnalysisReport) {
					if err := sender.SendAlert(context.Background(), ch, r); err != nil {
						log.Printf("Failed to send alert via %s (%s): %v ", ch.Name, ch.Type, err)
					}
				}(channel, report)
			}
		}
	}
}
