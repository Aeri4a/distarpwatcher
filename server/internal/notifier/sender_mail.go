package notifier

import (
	"context"
	"fmt"
	"log"

	"server/internal/analyzer"
	"server/internal/config"
	"server/internal/database"

	"github.com/wneessen/go-mail"
)

type MailSender struct {
	config config.MailConfig
}

func (sen *MailSender) SendAlert(ctx context.Context, channel database.NotificationChannel, alert *analyzer.Alert) error {
	m := mail.NewMsg()

	if err := m.From(sen.config.From); err != nil {
		return fmt.Errorf("failed to set From address: %w", err)
	}

	if err := m.To(channel.Target); err != nil {
		return fmt.Errorf("failed to set To address: %w", err)
	}

	m.Subject(fmt.Sprintf("Security Alert: %s detected on Agent %s", alert.AttackType, alert.AgentId))

	body := fmt.Sprintf(`
		<html>
		<body style="font-family: Arial, sans-serif; color: #333;">
			<h2 style="color: #d9534f;">Distributed ARP Watcher Alert</h2>
			<hr>
			<p><strong>Attack Type:</strong> %s</p>
			<p><strong>Target IP:</strong> %s</p>
			<p><strong>Agent ID:</strong> %s</p>
			<br>
			<h3>Message:</h3>
			<p style="background-color: #f9f9f9; padding: 10px; border-left: 4px solid #d9534f;">
				%s
			</p>
		</body>
		</html>
	`, alert.AttackType, alert.TargetIp, alert.AgentId, alert.Message)

	m.SetBodyString(mail.TypeTextHTML, body)

	var options []mail.Option
	options = append(options, mail.WithPort(sen.config.Port))

	if sen.config.Username != "" && sen.config.Password != "" {
		options = append(options, mail.WithSMTPAuth(mail.SMTPAuthPlain))
		options = append(options, mail.WithUsername(sen.config.Username))
		options = append(options, mail.WithPassword(sen.config.Password))
	} else {
		// for local testing
		options = append(options, mail.WithTLSPolicy(mail.NoTLS))
	}

	client, err := mail.NewClient(sen.config.Host, options...)
	if err != nil {
		return fmt.Errorf("failed to create mail client: %w", err)
	}

	if err := client.DialAndSendWithContext(ctx, m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	log.Printf("[Sender][MAIL] Alert sent successfully to [%s], by channel name %s", channel.Target, channel.Name)
	return nil
}
