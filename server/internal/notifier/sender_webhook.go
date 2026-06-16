package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"server/internal/analyzer"
	"server/internal/database"
)

type WebhookSender struct{}

type WebhookPayload struct {
	AgentId  string `json:"agent_id"`
	TargetIp string `json:"target_ip"`
	Attack   string `json:"attack"`
	Message  string `json:"message"`
}

func (sen *WebhookSender) SendAlert(ctx context.Context, channel database.NotificationChannel, alert *analyzer.Alert) error {
	url := channel.Target

	payload := WebhookPayload{
		AgentId:  alert.AgentId,
		TargetIp: alert.TargetIp,
		Attack:   string(alert.AttackType),
		Message:  alert.Message,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	log.Printf("[Sender][Webhook] Alert sent successfully to [%s], by channel name %s", channel.Target, channel.Name)

	return nil
}
