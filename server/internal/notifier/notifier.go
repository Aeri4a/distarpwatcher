package notifier

import (
	"context"
	"fmt"
	"log"
	"server/internal/analyzer"
	"server/internal/config"
	"server/internal/database"
	"time"
)

type Sender interface {
	SendAlert(ctx context.Context, channel database.NotificationChannel, alert *analyzer.Alert) error
}

type Notifier struct {
	db               database.Database
	notificationChan chan *analyzer.Alert

	senders map[string]Sender

	lastAlerted    map[string]time.Time
	cooldownPeriod time.Duration
}

func NewNotifier(cfg config.Config, db database.Database, notificationChan chan *analyzer.Alert) *Notifier {
	notif := &Notifier{
		db:               db,
		notificationChan: notificationChan,
		senders:          make(map[string]Sender),
		lastAlerted:      make(map[string]time.Time),
		cooldownPeriod:   time.Minute * 2,
	}

	notif.senders["MAIL"] = &MailSender{config: cfg.Mail}
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
		case alert, ok := <-notif.notificationChan:
			if !ok {
				log.Println("Notification channel closed.")
				return nil
			}

			if notif.shouldThrottle(alert) {
				log.Println("Alert throttled.")
				continue
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

				go func(ch database.NotificationChannel, al *analyzer.Alert) {
					if err := sender.SendAlert(context.Background(), ch, al); err != nil {
						log.Printf("Failed to send alert via %s (%s): %v ", ch.Name, ch.Type, err)
					}
				}(channel, alert)
			}
		}
	}
}

func (notif *Notifier) shouldThrottle(alert *analyzer.Alert) bool {
	// there should be some cleaning for lastAlerted
	signature := fmt.Sprintf("%s:%s:%s", alert.AgentId, alert.TargetIp, alert.AttackType)

	lastAlerted, exist := notif.lastAlerted[signature]
	if !exist {
		notif.lastAlerted[signature] = time.Now()
		return false
	}

	if time.Since(lastAlerted) < notif.cooldownPeriod {
		return true
	}

	notif.lastAlerted[signature] = time.Now()
	return false
}
